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

func (d *DiscordMessage) Parse() (*core.Message, error) {
	msg := parse(d.message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessage) Scope(type_ int) (int64, error) {
	return getScope(type_, d.message.ChannelID, d.message.GuildID, d.message.Author.ID)
}

func (d *DiscordMessage) Write(msg interface{}, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.session, msg.(string), d.message.ChannelID)

	case *dg.MessageEmbed:
		// TODO
		return nil, nil
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}
