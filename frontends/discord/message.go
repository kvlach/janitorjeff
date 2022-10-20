package discord

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

type DiscordMessage struct {
	session *dg.Session
	message *dg.Message
}

func (d *DiscordMessage) Admin() bool {
	return isAdmin(d.message.Author.ID)
}

func (d *DiscordMessage) Parse() (*core.Message, error) {
	msg := parse(d.message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessage) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.session, d.message)
}

func (d *DiscordMessage) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.session, d.message)
}

func (d *DiscordMessage) PersonScope(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *DiscordMessage) PlaceScope(id string) (int64, error) {
	return getPlaceScope(id, d.message, d.session)
}

func (d *DiscordMessage) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.session, msg.(string), d.message.ChannelID, d.message.GuildID)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.session, d.message, embed, usrErr)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}