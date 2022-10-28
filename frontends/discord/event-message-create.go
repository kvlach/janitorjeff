package discord

import (
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type DiscordMessageCreate struct {
	Session *dg.Session
	Message *dg.MessageCreate
}

func (d *DiscordMessageCreate) Admin() bool {
	return isAdmin(d.Message.Author.ID)
}

func (d *DiscordMessageCreate) Parse() (*core.Message, error) {
	msg := parse(d.Message.Message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessageCreate) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Session, d.Message.Message)
}

func (d *DiscordMessageCreate) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.Session, d.Message.Message)
}

func (d *DiscordMessageCreate) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *DiscordMessageCreate) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.Message.Message, d.Session)
}

func (d *DiscordMessageCreate) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.Message.Message, d.Session)
}

func (d *DiscordMessageCreate) ReplyUsage(usage string) any {
	return replyUsage(usage)
}

func (d *DiscordMessageCreate) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Session, msg.(string), d.Message.ChannelID, d.Message.GuildID)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.Session, d.Message.Message, embed, usrErr)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

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

	// TODO: remove this when each server can configure which commands will be
	// active
	if m.GuildID == "348368013382254602" && !strings.HasPrefix(m.Content, "!pb") {
		return
	}

	d := &DiscordMessageCreate{s, m}
	msg, err := d.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	msg.Run()
}
