package discord

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/utils"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var errInvalidID = errors.New("given string is not a valid ID")

func getChannelGuildIDs(id string, m *dg.Message, s *dg.Session) (string, string, error) {
	var channelID, guildID string

	if id == m.ChannelID || id == m.GuildID {
		channelID = m.ChannelID
		guildID = m.GuildID
	} else if channel, err := s.Channel(id); err == nil {
		channelID = channel.ID
		guildID = channel.GuildID
	} else if guild, err := s.Guild(id); err == nil {
		channelID = ""
		guildID = guild.ID
	} else {
		return "", "", fmt.Errorf("id '%s' not guild or channel id", id)
	}

	return channelID, guildID, nil
}

func getGuildScope(tx *sql.Tx, id string) (int64, error) {
	if guild, err := dbGetGuildScope(id); err == nil {
		return guild, nil
	}
	return dbAddGuildScope(tx, id)
}

func getChannelScope(tx *sql.Tx, id string, guild int64) (int64, error) {
	if channel, err := dbGetChannelScope(id); err == nil {
		return channel, nil
	}
	return dbAddChannelScope(tx, id, guild)
}

func getPlaceExactScope(id string, m *dg.Message, s *dg.Session) (int64, error) {
	channelID, guildID, err := getChannelGuildIDs(id, m, s)
	if err != nil {
		return -1, err
	}

	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	guild, err := getGuildScope(tx, guildID)
	if err != nil {
		return -1, err
	}

	if id == guildID {
		return guild, nil
	}

	return getChannelScope(tx, channelID, guild)
}

func getPlaceLogicalScope(id string, m *dg.Message, s *dg.Session) (int64, error) {
	channelID, guildID, err := getChannelGuildIDs(id, m, s)
	if err != nil {
		return -1, err
	}

	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	if channelScope, err := dbGetChannelScope(channelID); err == nil {
		if guildID == "" {
			return channelScope, nil
		}

		// Find channel's guild scope. If the channel scope exists then guild
		// one does also, even if it's the special empty guild
		return dbGetGuildFromChannel(channelScope)
	}

	// only create a new guildScope if it doesn't already exist
	guildScope, err := getGuildScope(tx, guildID)
	if err != nil {
		return -1, err
	}

	// in case the passed id is an arbitrary guild, which means that we can't
	// deduce a specific channel
	if channelID == "" {
		return guildScope, nil
	}

	if guildID != "" {
		return guildScope, nil
	}

	return dbAddChannelScope(tx, channelID, guildScope)
}

func getPersonScope(id string) (int64, error) {
	// doesn't check if the ID is valid, that job is handled by the PersonID
	// function
	scope, err := dbGetUserScope(id)
	if err == nil {
		return scope, nil
	}

	db := core.Globals.DB

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	scope, err = dbAddUserScope(tx, id)
	if err != nil {
		return -1, err
	}

	return scope, tx.Commit()
}

func replyUsage(usage string) *dg.MessageEmbed {
	embed := &dg.MessageEmbed{
		Title: fmt.Sprintf("Usage: `%s`", usage),
	}
	return embed
}

func getPersonID(s, guildID string, ds *dg.Session, msg *dg.Message) (string, error) {
	// expected inputs are either the id itself or a mention which looks like
	// this: <@id>

	// trim them if they exist
	s = strings.TrimPrefix(s, "<@")
	s = strings.TrimSuffix(s, ">")

	// not even a number, no point in asking discord if it's a valid id
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return "", errInvalidID
	}

	// if the user is in a DM then the only valid user id would be that of
	// themselves, so having no guild isn't a problem
	if s == msg.Author.ID {
		return s, nil
	}

	if _, err := ds.GuildMember(guildID, s); err != nil {
		return "", err
	}

	return s, nil
}

func getPlaceID(s string, ds *dg.Session, msg *dg.Message) (string, error) {
	// expects one of the following:
	// - guild id
	// - channel id
	// - channel mention (looks like: <#channel-id>)

	s = strings.TrimPrefix(s, "<#")
	s = strings.TrimSuffix(s, ">")

	// if it's not a number there's no point in asking discord if it's a valid
	// id
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return s, errInvalidID
	}

	// will usually be a guild, so first try checking that
	if _, err := ds.Guild(s); err == nil {
		return s, nil
	}

	if _, err := ds.Channel(s); err == nil {
		return s, nil
	}

	return s, errInvalidID
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
		Type: Type,
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

func sendText(d *dg.Session, text, channel, guild string) (*core.Message, error) {
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

	// the returned response leaves the guild id empty whether or not the
	// message was sent to a guild, this makes it very difficult for the scope
	// getter functions to know which scope to return (they check if the guild
	// id is empty to know whether or not a message came from a DM), so we
	// manually set the guild id here
	msg.GuildID = guild

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

	// the returned response leaves the guild id empty whether or not the
	// message was sent to a guild, this makes it very difficult for the scope
	// getter functions to know which scope to return (they check if the guild
	// id is empty to know whether or not a message came from a DM), so we
	// manually set the guild id here
	resp.GuildID = m.GuildID

	replies.Set(m.ID, resp.ID)
	return (&DiscordMessage{d, resp}).Parse()
}

func editEmbed(d *dg.Session, m *dg.Message, embed *dg.MessageEmbed, usrErr error, id string) (*core.Message, error) {
	embed = embedColor(embed, usrErr)

	msgEdit := &dg.MessageEdit{
		ID:      id,
		Channel: m.ChannelID,

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

	// the returned response leaves the guild id empty whether or not the
	// message was sent to a guild, this makes it very difficult for the scope
	// getter functions to know which scope to return (they check if the guild
	// id is empty to know whether or not a message came from a DM), so we
	// manually set the guild id here
	resp.GuildID = m.GuildID

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
