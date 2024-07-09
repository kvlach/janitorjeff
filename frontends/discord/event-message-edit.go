package discord

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

const rdbMessageReplyToKeyPrefix = "frontend_discord_reply_"

type MessageEdit struct {
	Message *dg.MessageUpdate
	VC      *dg.VoiceConnection
}

func messageEdit(_ *dg.Session, m *dg.MessageUpdate) {
	// For some reason this randomly gets triggered if a link has been sent
	// and there has been no edit to the message. Perhaps it could be have
	// something to do with the message getting automatically edited in order
	// for the embed to added. This is difficult to test as it usually doesn't
	// happen.
	if m.Author == nil {
		return
	}

	if m.Author.Bot {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	d := &MessageEdit{
		Message: m,
	}
	msg, err := d.Parse()
	if err != nil {
		return
	}
	core.EventMessageChan <- msg
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (d *MessageEdit) Parse() (*core.EventMessage, error) {
	msg := parse(d.Message.Message)
	msg.Client = d
	return msg, nil
}

func (d *MessageEdit) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID)
}

func (d *MessageEdit) PlaceID(s string) (string, error) {
	return getPlaceID(s)
}

func (d *MessageEdit) Person(id string) (int64, error) {
	return dbGetPersonScope(id)
}

func (d *MessageEdit) send(msg any, urr error, ping bool) (*core.EventMessage, error) {
	rdbKey := rdbMessageReplyToKeyPrefix + d.Message.ID

	switch t := msg.(type) {
	case string:
		text := msg.(string)
		id, err := core.RDB.Get(ctx, rdbKey).Result()
		if err != nil {
			return sendText(d.Message.Message, text, ping)
		}
		return editText(d.Message.Message, id, text)

	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		id, err := core.RDB.Get(ctx, rdbKey).Result()
		if err != nil {
			return sendEmbed(d.Message.Message, embed, urr, ping)
		}
		return editEmbed(d.Message.Message, embed, urr, id)

	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (d *MessageEdit) Send(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.send(msg, urr, false)
}

func (d *MessageEdit) Ping(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.send(msg, urr, true)
}

func (d *MessageEdit) Write(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.Send(msg, urr)
}

func (d *MessageEdit) Natural(msg any, urr core.Urr) (*core.EventMessage, error) {
	return d.Send(msg, urr)
}

func (d *MessageEdit) QuoteCommand(cmd string) string {
	return PlaceInBackticks(cmd)
}
