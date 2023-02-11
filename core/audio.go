package core

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"

	"github.com/janitorjeff/gosafe"
)

const (
	Play int = iota
	Loop
	Pause
	Stop
)

// A select with multiple ready cases chooses one pseudo-randomly. So if the
// goroutine is "slow" to check those channels, you might send a value on both
// pause and resume (assuming they are buffered) so receiving from both channels
// could be ready, and resume could be chosen first, and in a later iteration
// the pause when the goroutine should not be paused anymore.
//
// Source: https://stackoverflow.com/a/60490371
type State struct {
	gosafe.Value[int]
}

type Speaker interface {
	// Enabled returns true if the frontend supports voice chat that the bot can
	// connect to.
	Enabled() bool

	// The audio's expected frame rate.
	FrameRate() int

	// The audio's expected number of channels.
	Channels() int

	// Join the message author's voice channel, if they are not connected to
	// any then returns an error. If in a specific frontend only one voice
	// channel will ever exist then the user doesn't have to be connected to it
	// for the bot to join (for example a discord server with only one voice
	// channel would not apply here as other ones *could* be created at any
	// point).
	Join() error

	// Send audio. Must have connected to a voice channel first, otherwise
	// returns an error.
	Say(buf io.Reader, s *State) error
}

// FFmpegBufferPipe will pipe audio coming from a buffer into ffmpeg and
// transform into audio that the speaker can transmit.
func FFmpegBufferPipe(sp Speaker, inBuf io.ReadCloser, st *State) error {
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
	defer ffmpeg.Process.Kill()

	sp.Say(ffmpegbuf, st)
	return nil
}

// FFmpegCommandPipe works exactly like FFmpegBufferPipe except it accepts a
// command instead of a buffer. Provided just for convenience.
func FFmpegCommandPipe(sp Speaker, cmd *exec.Cmd, st *State) error {
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Process.Kill()
	return FFmpegBufferPipe(sp, pipe, st)
}
