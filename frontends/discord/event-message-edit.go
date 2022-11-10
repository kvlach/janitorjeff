package discord

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"git.slowtyper.com/slowtyper/gosafe"
	dg "github.com/bwmarrin/discordgo"
)

var replies = gosafe.Map[string, string]{}

type MessageEdit struct {
	Session *dg.Session
	Message *dg.MessageUpdate
}

func (d *MessageEdit) Admin() bool {
	return isAdmin(d.Message.Author.ID)
}

func (d *MessageEdit) Parse() (*core.Message, error) {
	msg := parse(d.Message.Message)
	msg.Client = d
	return msg, nil
}

func (d *MessageEdit) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID, d.Session)
}

func (d *MessageEdit) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.Session)
}

func (d *MessageEdit) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *MessageEdit) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.Message.ChannelID, d.Message.GuildID, d.Session)
}

func (d *MessageEdit) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.Message.ChannelID, d.Message.GuildID, d.Session)
}

func (d *MessageEdit) Usage(usage string) any {
	return getUsage(usage)
}

func (d *MessageEdit) send(msg any, usrErr error, ping bool) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		text := msg.(string)
		id, ok := replies.Get(d.Message.ID)
		if !ok {
			return sendText(d.Session, d.Message.Message, text, ping)
		}
		return editText(d.Session, d.Message.Message, id, text)

	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		id, ok := replies.Get(d.Message.ID)
		if !ok {
			return sendEmbed(d.Session, d.Message.Message, embed, usrErr, ping)
		}
		return editEmbed(d.Session, d.Message.Message, embed, usrErr, id)

	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (d *MessageEdit) Send(msg any, usrErr error) (*core.Message, error) {
	return d.send(msg, usrErr, false)
}

func (d *MessageEdit) Ping(msg any, usrErr error) (*core.Message, error) {
	return d.send(msg, usrErr, true)
}

func (d *MessageEdit) Write(msg any, usrErr error) (*core.Message, error) {
	return d.Send(msg, usrErr)
}

func messageEdit(s *dg.Session, m *dg.MessageUpdate) {
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

	d := &MessageEdit{s, m}
	msg, err := d.Parse()
	if err != nil {
		return
	}
	msg.Run()
}
