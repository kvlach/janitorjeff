package discord

import (
	"fmt"
	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type MessageCreate struct {
	Message *dg.MessageCreate
	VC      *dg.VoiceConnection
}

func messageCreate(s *dg.Session, m *dg.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	d := &MessageCreate{
		Message: m,
	}
	msg, err := NewMessage(m.Message, d)
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}
	msg.Send()
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (d *MessageCreate) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID)
}

func (d *MessageCreate) PlaceID(s string) (string, error) {
	return getPlaceID(s)
}

func (d *MessageCreate) Person(id string) (int64, error) {
	return dbGetPersonScope(id)
}

func (d *MessageCreate) send(msg any, urr error, ping bool) (*core.EventMessage, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Message.Message, msg.(string), ping)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.Message.Message, embed, urr, ping)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (d *MessageCreate) Send(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.send(msg, urr, false)
}

func (d *MessageCreate) Ping(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.send(msg, urr, true)
}

func (d *MessageCreate) Write(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.Send(msg, urr)
}

func (d *MessageCreate) Natural(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.Send(msg, urr)
}

func (d *MessageCreate) QuoteCommand(cmd string) string {
	return PlaceInBackticks(cmd)
}
