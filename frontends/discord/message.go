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

func CreateClient(author, channel int64) (*DiscordMessage, error) {
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

	d := &DiscordMessage{
		session: core.Globals.Discord.Client,
		message: &dg.Message{
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

func (d *DiscordMessage) Admin() bool {
	return isAdmin(d.message.Author.ID)
}

func (d *DiscordMessage) Parse() (*core.Message, error) {
	msg := parse(d.message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessage) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.message.Author.ID, d.session)
}

func (d *DiscordMessage) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.session)
}

func (d *DiscordMessage) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *DiscordMessage) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.message.ChannelID, d.message.GuildID, d.session)
}

func (d *DiscordMessage) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.message.ChannelID, d.message.GuildID, d.session)
}

func (d *DiscordMessage) Usage(usage string) any {
	return getUsage(usage)
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
