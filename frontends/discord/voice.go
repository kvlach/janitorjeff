package discord

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/janitorjeff/jeff-bot/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"layeh.com/gopus"
)

type Speaker struct {
	GuildID  string
	AuthorID string
	VC       *dg.VoiceConnection
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
	v, err := joinUserVoiceChannel(sp.GuildID, sp.AuthorID)
	if err != nil {
		return err
	}
	sp.VC = v
	return nil
}

func (sp *Speaker) Say(buf io.Reader, s *core.AudioState) error {
	return voicePlay(sp.VC, buf, s)
}

func (sp *Speaker) AuthorDeafened() (bool, error) {
	// THERE IS CURRENTLY NO WAY TO FETCH SOMEONE'S VOICE STATE THROUGH THE
	// REST API, SO IF IT'S NOT IN THE CACHE WE CAN GO FUCK OURSELVES I GUESS
	vs, err := Session.State.VoiceState(sp.GuildID, sp.AuthorID)
	if err != nil {
		log.Debug().
			Err(err).
			Str("guild", sp.GuildID).
			Str("user", sp.AuthorID).
			Msg("failed to get voice state")
		return false, err
	}
	return vs.SelfDeaf, nil
}

func (sp *Speaker) AuthorConnected() (bool, error) {
	// FIXME: If the user is connected to a different channel then this will
	// return true, we don't want this.

	// THERE IS CURRENTLY NO WAY TO FETCH SOMEONE'S VOICE STATE THROUGH THE
	// REST API, SO IF IT'S NOT IN THE CACHE WE CAN GO FUCK OURSELVES I GUESS
	_, err := Session.State.VoiceState(sp.GuildID, sp.AuthorID)
	if err != nil {
		log.Debug().
			Err(err).
			Str("guild", sp.GuildID).
			Str("user", sp.AuthorID).
			Msg("failed to get voice state")
	}
	// if no error, then a voice state exists
	return err == nil, nil
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

func findUserVoiceState(guildID, userID string) (*dg.VoiceState, error) {
	guild, err := Session.State.Guild(guildID)
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

// joinUserVoiceChannel joins a session to the same channel as another user.
func joinUserVoiceChannel(guildID, userID string) (*dg.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := findUserVoiceState(guildID, userID)
	if err != nil {
		return nil, err
	}

	// Join the user's channel and start unmuted and deafened.
	return Session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}

// sendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
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

func play(v *dg.VoiceConnection, buf io.Reader, s *core.AudioState) {
	if err := v.Speaking(true); err != nil {
		log.Debug().Err(err).Msg("Couldn't set speaking to true")
	}

	defer func() {
		if err := v.Speaking(false); err != nil {
			log.Debug().Err(err).Msg("Couldn't set speakig to false")
		}
	}()

	send := make(chan []int16, 2)
	defer close(send)

	exit := make(chan bool)
	go func() {
		sendPCM(v, send)
		exit <- true
	}()

	for {
		switch s.Get() {
		case core.AudioPlay, core.AudioLoop:
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

func voicePlay(v *dg.VoiceConnection, buf io.Reader, s *core.AudioState) error {
	if v == nil {
		return errors.New("not connected to a voice channel")
	}
	play(v, buf, s)
	return nil
}
