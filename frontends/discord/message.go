package discord

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

type Message struct {
	Session *dg.Session
	Message *dg.Message
}

func CreateClient(author, channel int64, msgID string) (*Message, error) {
	channelID, err := core.Globals.DB.ScopeID(channel)
	if err != nil {
		return nil, err
	}

	guild, err := dbGetGuildFromChannel(channel)
	if err != nil {
		return nil, err
	}

	guildID, err := core.Globals.DB.ScopeID(guild)
	if err != nil {
		return nil, err
	}

	authorID, err := core.Globals.DB.ScopeID(author)
	if err != nil {
		return nil, err
	}

	d := &Message{
		Session: core.Globals.Discord.Client,
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

func (d *Message) Admin() bool {
	return isAdmin(d.Message.Author.ID)
}

func (d *Message) Parse() (*core.Message, error) {
	msg := parse(d.Message)
	msg.Client = d
	return msg, nil
}

func (d *Message) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID, d.Session)
}

func (d *Message) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.Session)
}

func (d *Message) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *Message) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.Message.ChannelID, d.Message.GuildID, d.Session)
}

func (d *Message) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.Message.ChannelID, d.Message.GuildID, d.Session)
}

func (d *Message) Usage(usage string) any {
	return getUsage(usage)
}

func (d *Message) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Session, d.Message, msg.(string))
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.Session, d.Message, embed, usrErr)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}
