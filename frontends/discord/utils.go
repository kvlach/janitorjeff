package discord

import (
	"fmt"
	"strconv"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/utils"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func getUserID(s string, ds *dg.Session, msg *dg.Message) (string, error) {
	// expected inputs are either the id itself or a mention which looks like
	// this: <@id>

	// trim them if they exist
	s = strings.TrimPrefix(s, "<@")
	s = strings.TrimSuffix(s, ">")

	// not even a number, no point in asking discord if it's a valid id
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return "", err
	}

	// if the user is in a DM then the only valid user id would be that of
	// themselves, so having no guild isn't a problem
	if s == msg.Author.ID {
		return s, nil
	}

	if _, err := ds.GuildMember(msg.GuildID, s); err != nil {
		return "", err
	}

	return s, nil
}

func getID(t int, s string, ds *dg.Session, msg *dg.Message) (string, error) {
	switch t {
	case User:
		return getUserID(s, ds, msg)
	default:
		return "", fmt.Errorf("id extraction for type '%d' not supported", t)
	}
}

func isAdmin(id string) bool {
	for _, admin := range core.Globals.Discord.Admins {
		if id == admin {
			return true
		}
	}
	return false
}

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

	embed = embedColor(embed, usrErr)

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
	replies.Set(m.ID, resp.ID)
	return (&DiscordMessage{d, resp}).Parse()
}

func editEmbed(d *dg.Session, embed *dg.MessageEmbed, usrErr error, id, channel string) (*core.Message, error) {
	embed = embedColor(embed, usrErr)

	msgEdit := &dg.MessageEdit{
		ID:      id,
		Channel: channel,

		Embeds: []*dg.MessageEmbed{
			embed,
		},
		AllowedMentions: &dg.MessageAllowedMentions{
			Parse: []dg.AllowedMentionType{}, // don't ping user
		},
	}

	resp, err := d.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return nil, err
	}
	return (&DiscordMessage{d, resp}).Parse()
}

func embedColor(embed *dg.MessageEmbed, usrErr error) *dg.MessageEmbed {
	if embed.Color != 0 {
		return embed
	}

	// default value of EmbedColor is 0 so even if it's not been set
	// then everything should be ok
	if usrErr == nil {
		embed.Color = core.Globals.Discord.EmbedColor
	} else {
		embed.Color = core.Globals.Discord.EmbedErrColor
	}
	return embed
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
