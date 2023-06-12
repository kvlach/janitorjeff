package core

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

type AudioState int

const (
	AudioPlay AudioState = iota
	AudioLoopAll
	AudioLoopCurrent
	AudioPause
	AudioStop
	AudioSkip
	AudioSeek
	AudioShuffle
)

type AudioSpeaker interface {
	// Enabled returns true if the frontend supports voice chat that the bot can
	// connect to.
	Enabled() bool

	// FrameRate returns the audio's expected frame rate. Is passed to ffmpeg
	// to convert the audio stream to the correct format before sending.
	FrameRate() int

	// Channels returns the audio's expected number of channels. Is passed to
	// ffmpeg to convert the audio stream to the correct format before sending.
	Channels() int

	// Join the message author's voice channel, if they are not connected to
	// any then returns an error. If in a specific frontend only one voice
	// channel will ever exist then the user doesn't have to be connected to it
	// for the bot to join (for example a discord server with only one voice
	// channel would not apply here as other ones *could* be created at any
	// point).
	Join() error

	// Say sends audio. Must have connected to a voice channel first, otherwise
	// returns an error.
	Say(buf io.Reader, st <-chan AudioState) error

	// AuthorDeafened returns true if the author that originally made the bot
	// join the voice channel is currently deafened.
	AuthorDeafened() (bool, error)

	// AuthorConnected returns true if the author that originally made the bot
	// join the voice channel is currently connected to that same voice channel.
	AuthorConnected() (bool, error)
}

type AudioPlayer[T any] struct {
	queue        []T
	lock         sync.Mutex
	state        AudioState
	stateQueue   chan AudioState
	stateCurrent chan AudioState
	play         func(T, <-chan AudioState) error
	loopAll      bool
	loopCurrent  bool
	index        int
}

func (p *AudioPlayer[T]) Queue() []T {
	p.lock.Lock()
	defer p.lock.Unlock()
	tmp := make([]T, len(p.queue))
	copy(tmp, p.queue)
	return tmp
}

func (p *AudioPlayer[T]) Append(item T) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.queue = append(p.queue, item)
}

func (p *AudioPlayer[T]) Next() (T, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.queue) == 0 {
		var none T
		return none, true
	}

	if p.loopCurrent {
		return p.queue[0], false
	}

	if p.loopAll {
		p.index++
		if p.index == len(p.queue) {
			p.index = 0
		}
		next := p.queue[p.index]
		return next, false
	}

	next := p.queue[0]
	p.queue = append([]T{}, p.queue[1:]...)
	return next, false
}

func (p *AudioPlayer[T]) Start() {
	p.stateQueue = make(chan AudioState)

	go func() {
		if p.play == nil {
			log.Debug().Msg("received play event without having a handler for it")
			return
		}

		for {
			current, skip := p.Next()
			if skip {
				continue
			}
			p.stateCurrent = make(chan AudioState)
			if err := p.play(current, p.stateCurrent); err != nil {
				log.Error().Err(err).Msg("failed to play item")
			}
		}
	}()

	for {
		switch p.state = <-p.stateQueue; p.state {
		case AudioPlay:
			p.stateCurrent <- AudioPlay

		case AudioLoopAll:
			// TODO: check if loop current is on
			p.loopAll = true

		case AudioLoopCurrent:
			// TODO: check if loop all is on
			p.loopCurrent = true

		case AudioPause:
			log.Debug().Msg("received pause event")
			p.stateCurrent <- AudioPause

		case AudioStop:
			log.Debug().Msg("received stop event")
			p.stateCurrent <- AudioStop
			return

		case AudioSkip:
			p.stateCurrent <- AudioStop
		}
	}
}

func (p *AudioPlayer[T]) Current() AudioState {
	return p.state
}

func (p *AudioPlayer[T]) Play() {
	p.stateQueue <- AudioPlay
}

func (p *AudioPlayer[T]) Pause() {
	p.stateQueue <- AudioPause
}

func (p *AudioPlayer[T]) Stop() {
	p.stateQueue <- AudioStop
}

func (p *AudioPlayer[T]) LoopAll() {
	p.stateQueue <- AudioLoopAll
}

func (p *AudioPlayer[T]) LoopCurrent() {
	p.stateQueue <- AudioLoopCurrent
}

func (p *AudioPlayer[T]) Skip() {
	p.stateQueue <- AudioSkip
}

func (p *AudioPlayer[T]) HandlePlay(handler func(T, <-chan AudioState) error) {
	p.play = handler
}

// AudioProcessBuffer will pipe audio coming from a buffer into ffmpeg and
// transform into audio that the speaker can transmit.
func AudioProcessBuffer(sp AudioSpeaker, inBuf io.ReadCloser, st <-chan AudioState) error {
	ffmpeg := exec.Command(
		"ffmpeg",
		"-i", "-",
		"-f", "s16le",
		"-ar", strconv.Itoa(sp.FrameRate()),
		"-ac", strconv.Itoa(sp.Channels()),
		"pipe:1",
	)

	var err error
	ffmpeg.Stdin = inBuf

	ffmpegout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return err
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384) // 2**14

	if err := ffmpeg.Start(); err != nil {
		return err
	}
	defer func() {
		if err := ffmpeg.Process.Kill(); err != nil {
			log.Error().Err(err).Msg("failed to kill ffmpeg process")
		}
	}()
	return sp.Say(ffmpegbuf, st)
}

// AudioProcessCommand works exactly like AudioProcessBuffer except it accepts
// a command instead of a buffer. Provided just for convenience.
func AudioProcessCommand(sp AudioSpeaker, cmd *exec.Cmd, st <-chan AudioState) error {
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Error().Err(err).Msg("failed to kill command process")
		}
	}()
	return AudioProcessBuffer(sp, pipe, st)
}

// AudioProcessBytes works exactly like AudioProcessBuffer except it accepts a
// slice of bytes instead of a buffer. Provided just for convenience.
func AudioProcessBytes(sp AudioSpeaker, b []byte, st <-chan AudioState) error {
	buf := io.NopCloser(bytes.NewReader(b))
	return AudioProcessBuffer(sp, buf, st)
}
