package core

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strconv"

	"git.sr.ht/~slowtyper/gosafe"
	"github.com/rs/zerolog/log"
)

const (
	AudioPlay int = iota
	AudioLoop
	AudioPause
	AudioStop
)

// AudioState represents the current state of the audio (playing, paused, etc.).
// A select with multiple cases isn't used because it chooses one pseudo-randomly.
// So if the  goroutine is "slow" to check those channels, you might send a
// value on both pause and resume (assuming they are buffered) so receiving from
// both channels  could be ready, and resume could be chosen first, and in a
// later iteration the pause when the goroutine should not be paused anymore.
//
// Source: https://stackoverflow.com/a/60490371
type AudioState struct {
	gosafe.Value[int]
}

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
	Say(buf io.Reader, s *AudioState) error

	// AuthorDeafened returns true if the author that originally made the bot
	// join the voice channel is currently deafened.
	AuthorDeafened() (bool, error)

	// AuthorConnected returns true if the author that originally made the bot
	// join the voice channel is currently connected to that same voice channel.
	AuthorConnected() (bool, error)
}

// AudioProcessBuffer will pipe audio coming from a buffer into ffmpeg and
// transform into audio that the speaker can transmit.
func AudioProcessBuffer(sp AudioSpeaker, inBuf io.ReadCloser, st *AudioState) error {
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
func AudioProcessCommand(sp AudioSpeaker, cmd *exec.Cmd, st *AudioState) error {
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
func AudioProcessBytes(sp AudioSpeaker, b []byte, st *AudioState) error {
	buf := io.NopCloser(bytes.NewReader(b))
	return AudioProcessBuffer(sp, buf, st)
}
