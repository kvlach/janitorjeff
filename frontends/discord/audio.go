package discord

import (
	"encoding/binary"
	"errors"
	"io"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"layeh.com/gopus"
)

type Speaker struct {
	Author core.Personifier
	Here   core.Placer
	VC     *dg.VoiceConnection
}

func (*Speaker) Enabled() bool {
	return true
}

func (*Speaker) FrameRate() int {
	return frameRate
}

func (*Speaker) Channels() int {
	return channels
}

func (sp *Speaker) Join() error {
	aid, err := sp.Author.ID()
	if err != nil {
		return err
	}
	v, err := Client.VoiceJoin(sp.Here.IDLogical(), aid)
	if err != nil {
		return err
	}
	sp.VC = v
	return nil
}

func (sp *Speaker) Leave() error {
	if sp.VC == nil {
		return errors.New("not connected, can't disconnect")
	}
	if err := sp.VC.Disconnect(); err != nil {
		return err
	}
	sp.VC = nil
	return nil
}

func (sp *Speaker) Say(buf io.Reader, s <-chan core.AudioState) error {
	return voicePlay(sp.VC, buf, s)
}

func (sp *Speaker) AuthorDeafened() (bool, error) {
	aid, err := sp.Author.ID()
	if err != nil {
		return false, err
	}
	gid := sp.Here.IDLogical()

	vs, err := Client.VoiceState(gid, aid)
	if err != nil {
		log.Debug().
			Err(err).
			Str("guild", gid).
			Str("user", aid).
			Msg("failed to get voice state")
		return false, err
	}
	return vs.SelfDeaf, nil
}

func (sp *Speaker) AuthorConnected() (bool, error) {
	aid, err := sp.Author.ID()
	if err != nil {
		return false, err
	}
	vs, err := Client.VoiceState(sp.Here.IDLogical(), aid)
	// if error then no voice state exists, which means that the author is not
	// connected
	if err != nil {
		return false, nil
	}
	if sp.VC == nil {
		return false, errors.New("unexpected nil voice connection")
	}
	return vs.ChannelID == sp.VC.ChannelID, nil
}

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

// sendPCM will receive on the provided channel, encode
// received PCM data into Opus then send that to DiscordGo
func sendPCM(v *dg.VoiceConnection, pcm <-chan []int16) {
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

func play(v *dg.VoiceConnection, buf io.Reader, st <-chan core.AudioState) {
	if err := v.Speaking(true); err != nil {
		log.Debug().Err(err).Msg("Couldn't set speaking to true")
	}

	defer func() {
		if err := v.Speaking(false); err != nil {
			log.Debug().Err(err).Msg("Couldn't set speaking to false")
		}
	}()

	send := make(chan []int16, 2)
	defer close(send)

	exit := make(chan bool)
	go func() {
		sendPCM(v, send)
		exit <- true
	}()

	state := core.AudioPlay
	go func() {
		for {
			state = <-st
		}
	}()

	for {
		switch state {
		case core.AudioPlay:
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
		case core.AudioPause:
		case core.AudioStop:
			return
		}
	}
}

func voicePlay(v *dg.VoiceConnection, buf io.Reader, s <-chan core.AudioState) error {
	if v == nil {
		return errors.New("not connected to a voice channel")
	}
	play(v, buf, s)
	return nil
}
