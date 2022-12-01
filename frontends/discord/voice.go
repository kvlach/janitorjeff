package discord

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"

	"github.com/janitorjeff/jeff-bot/core"

	dg "github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

// *Very* heavily inspired from https://github.com/bwmarrin/dgvoice/

// Technically the below settings can be adjusted however that poses
// a lot of other problems that are not handled well at this time.
// These below values seem to provide the best overall performance
const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

var opusEncoder *gopus.Encoder

func FindUserVoiceState(s *dg.Session, guildID, userID string) (*dg.VoiceState, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}
	return nil, errors.New("Could not find user's voice state")
}

// JoinUserVoiceChannel joins a session to the same channel as another user.
func JoinUserVoiceChannel(s *dg.Session, guildID, userID string) (*dg.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := FindUserVoiceState(s, guildID, userID)
	if err != nil {
		return nil, err
	}

	// Join the user's channel and start unmuted and deafened.
	return s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func SendPCM(v *dg.VoiceConnection, pcm <-chan []int16) {
	if pcm == nil {
		return
	}

	var err error
	opusEncoder, err = gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		return
	}

	for {
		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			return
		}

		// try encoding pcm frame with Opus
		opus, err := opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			return
		}
		// send encoded opus data to the sendOpus channel
		v.OpusSend <- opus
	}
}

func Play(v *dg.VoiceConnection, buf io.Reader, s *core.State) {
	send := make(chan []int16, 2)
	defer close(send)

	exit := make(chan bool)
	go func() {
		SendPCM(v, send)
		exit <- true
	}()

	for {
		switch s.Get() {
		case core.Play:
			audiobuf := make([]int16, frameSize*channels)
			err := binary.Read(buf, binary.LittleEndian, &audiobuf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return
			}

			select {
			case send <- audiobuf:
			case <-exit:
				return
			}
		case core.Pause:
		case core.Stop:
			return
		}
	}
}

func PipeThroughFFmpeg(v *dg.VoiceConnection, cmd *exec.Cmd, s *core.State) error {
	ffmpeg := exec.Command(
		"ffmpeg",
		"-i", "-",
		"-f", "s16le",
		"-ar", strconv.Itoa(frameRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	)

	var err error
	ffmpeg.Stdin, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}

	ffmpegout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return err
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384) // 2**14

	if err := ffmpeg.Start(); err != nil {
		return err
	}
	defer ffmpeg.Process.Kill()

	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Process.Kill()

	Play(v, ffmpegbuf, s)
	return nil
}