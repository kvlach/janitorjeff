package discord

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"

	dg "github.com/bwmarrin/discordgo"
)

type Message struct {
	Message *dg.Message
	VC      *dg.VoiceConnection
}

func CreateClient(author, channel int64, msgID string) (*Message, error) {
	channelID, err := core.DB.ScopeID(channel)
	if err != nil {
		return nil, err
	}

	guild, err := dbGetGuildFromChannel(channel)
	if err != nil {
		return nil, err
	}

	guildID, err := core.DB.ScopeID(guild)
	if err != nil {
		return nil, err
	}

	authorID, err := core.DB.ScopeID(author)
	if err != nil {
		return nil, err
	}

	// check if message id still exists (could have been deleted for example)
	if _, err := Session.ChannelMessage(channelID, msgID); err != nil {
		msgID = ""
	}

	d := &Message{
		Message: &dg.Message{
			ID:        msgID,
			ChannelID: channelID,
			GuildID:   guildID,
			Author: &dg.User{
				ID:  authorID,
				Bot: false,
			},
		},
	}

	return d, nil
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (d *Message) Parse() (*core.Message, error) {
	msg := parse(d.Message)
	msg.Client = d
	return msg, nil
}

func (d *Message) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID)
}

func (d *Message) PlaceID(s string) (string, error) {
	return getPlaceID(s)
}

func (d *Message) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *Message) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.Message.ChannelID, d.Message.GuildID)
}

func (d *Message) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.Message.ChannelID, d.Message.GuildID)
}

func (d *Message) Usage(usage string) any {
	return getUsage(usage)
}

func (d *Message) send(msg any, usrErr error, ping bool) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Message, msg.(string), ping)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.Message, embed, usrErr, ping)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (d *Message) Send(msg any, usrErr error) (*core.Message, error) {
	return d.send(msg, usrErr, false)
}

func (d *Message) Ping(msg any, usrErr error) (*core.Message, error) {
	return d.send(msg, usrErr, true)
}

func (d *Message) Write(msg any, usrErr error) (*core.Message, error) {
	return d.Send(msg, usrErr)
}
