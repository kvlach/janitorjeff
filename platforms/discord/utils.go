package discord

import (
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/utils"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func parse(m *dg.Message) *core.Message {
	log.Debug().
		Msg("starting to parse message")

	author := &core.Author{
		ID:          m.Author.ID,
		Name:        m.Author.Username,
		DisplayName: getDisplayName(m.Member, m.Author),
		Mention:     m.Author.Mention(),
	}

	channel := &core.Channel{
		ID:   m.ChannelID,
		Name: m.ChannelID,
	}

	msg := &core.Message{
		ID:   m.ID,
		Type: core.Discord,
		Raw:  m.Content,
		// GuildID is always empty in returned message objects, this is here in
		// case that changes in the future.
		IsDM:    m.GuildID == "",
		Author:  author,
		Channel: channel,
	}

	return msg
}

func getDisplayName(member *dg.Member, author *dg.User) string {
	var displayName string

	if member == nil || member.Nick == "" {
		displayName = author.Username
	} else {
		displayName = member.Nick
	}

	return displayName
}

func sendText(d *dg.Session, text, channel string) (*core.Message, error) {
	var msg *dg.Message
	var err error

	lenLim := 2000
	// TODO: grapheme clusters instead of plain len?
	lenCnt := func(s string) int { return len(s) }

	if lenLim > lenCnt(text) {
		msg, err = d.ChannelMessageSend(channel, text)
	} else {
		parts := utils.Split(text, lenCnt, lenLim)
		for _, p := range parts {
			msg, err = d.ChannelMessageSend(channel, p)
		}
	}

	if err != nil {
		return nil, err
	}
	m := &DiscordMessage{d, msg}
	return m.Parse()
}

func sendEmbed(d *dg.Session, m *dg.Message, embed *dg.MessageEmbed, usrErr error) (*core.Message, error) {
	// TODO: implement message scrolling
	if embed.Color == 0 {
		// default value of EmbedColor is 0 so even if it's not been set
		// then everything should be ok
		if usrErr == nil {
			embed.Color = core.Globals.Discord.EmbedColor
		} else {
			embed.Color = core.Globals.Discord.EmbedErrColor
		}
	}

	// TODO: Consider adding an option which allows one of these 3 values
	// - no reply + no ping, just an embed
	// - reply + no ping (default)
	// - reply + ping
	// Maybe even no embed and just plain text?
	msgSend := &dg.MessageSend{
		Embeds: []*dg.MessageEmbed{
			embed,
		},
		AllowedMentions: &dg.MessageAllowedMentions{
			Parse: []dg.AllowedMentionType{}, // don't ping user
		},
		Reference: m.Reference(),
	}

	resp, err := d.ChannelMessageSendComplex(m.ChannelID, msgSend)
	if err != nil {
		return nil, err
	}
	replies[m.ID] = resp.ID
	return (&DiscordMessage{d, resp}).Parse()
}

func PlaceInBackticks(s string) string {
	if !strings.Contains(s, "`") {
		return fmt.Sprintf("`%s`", s)
	}

	// Only way I could find to display backticks correctly. Very hacky.
	// Works for an arbitrary number of backticks. Works everywhere except
	// on android.

	const zeroWidthSpace = "\u200b"
	// zeroWidthSpace := "\u3164"

	s = strings.ReplaceAll(s, "`", zeroWidthSpace+"`"+zeroWidthSpace)
	return fmt.Sprintf("``%s``", s)
}
