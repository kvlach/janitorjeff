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

func (d *DiscordMessage) ID(t int, s string) (string, error) {
	return getID(t, s, d.session, d.message)
}

func (d *DiscordMessage) Scope(t int, id string) (int64, error) {
	return getScope(t, id, d.message)
}

func (d *DiscordMessage) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.session, msg.(string), d.message.ChannelID)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.session, d.message, embed, usrErr)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}
