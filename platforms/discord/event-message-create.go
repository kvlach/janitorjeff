package discord

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type DiscordMessageCreate struct {
	Session *dg.Session
	Message *dg.MessageCreate
}

func (d *DiscordMessageCreate) Parse() (*core.Message, error) {
	msg := parse(d.Message.Message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessageCreate) Scope(type_ int) (int64, error) {
	return getScope(type_, d.Message.ChannelID, d.Message.GuildID, d.Message.Author.ID)
}

func (d *DiscordMessageCreate) Write(msg interface{}, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Session, msg.(string), d.Message.ChannelID)
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

	d := &DiscordMessageCreate{s, m}
	msg, err := d.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	msg.Run()
}
