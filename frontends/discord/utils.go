package discord

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

var errInvalidID = errors.New("given string is not a valid ID")

func getChannelGuildIDs(id string) (string, string, error) {
	var channelID, guildID string

	if ch, err := Client.Channel(id); err == nil {
		channelID = ch.ID
		guildID = ch.GuildID
	} else if g, err := Client.Guild(id); err == nil {
		guildID = g.ID

		// Try to find a text channel in the guild
		if len(g.Channels) == 0 {
			return "", "", errors.New("expected to find channels in guild")
		}
		for _, ch := range g.Channels {
			if ch.Type == dg.ChannelTypeGuildText {
				channelID = ch.ID
				break
			}
		}
		if channelID == "" {
			return "", "", errors.New("couldn't find text channel")
		}
	} else {
		return "", "", fmt.Errorf("id '%s' not guild or channel id", id)
	}

	return channelID, guildID, nil
}

func getChannelGuildScopes(id string) (int64, int64, error) {
	channelID, guildID, err := getChannelGuildIDs(id)
	if err != nil {
		return 0, 0, err
	}

	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	gs, err := getGuildScope(tx, guildID)
	if err != nil {
		return 0, 0, err
	}
	cs, err := getChannelScope(tx, channelID, gs)
	if err != nil {
		return 0, 0, err
	}
	return cs, gs, tx.Commit()
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

func getPlaceExactScope(id string) (int64, error) {
	channelID, guildID, err := getChannelGuildIDs(id)
	if err != nil {
		return -1, err
	}

	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	guild, err := getGuildScope(tx, guildID)
	if err != nil {
		return -1, err
	}

	if id == guildID {
		return guild, nil
	}

	channel, err := getChannelScope(tx, channelID, guild)
	if err != nil {
		return -1, err
	}

	return channel, tx.Commit()
}

func getPlaceLogicalScope(id string) (int64, error) {
	channelID, guildID, err := getChannelGuildIDs(id)
	if err != nil {
		return -1, err
	}

	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	//goland:noinspection GoUnhandledErrorResult
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
		return guildScope, tx.Commit()
	}

	if guildID != "" {
		return guildScope, tx.Commit()
	}

	channelScope, err := dbAddChannelScope(tx, channelID, guildScope)
	if err != nil {
		return -1, err
	}

	return channelScope, tx.Commit()
}

func dbGetPersonScope(id string) (int64, error) {
	// doesn't check if the ID is valid, that job is handled by the PersonID
	// function
	scope, err := dbGetUserScope(id)
	if err == nil {
		return scope, nil
	}

	db := core.DB

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	scope, err = dbAddUserScope(tx, id)
	if err != nil {
		return -1, err
	}

	return scope, tx.Commit()
}

func getPersonID(s, guildID, authorID string) (string, error) {
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
	if s == authorID {
		return s, nil
	}

	if _, err := Client.Member(guildID, s); err != nil {
		return "", err
	}

	return s, nil
}

func getPlaceID(s string) (string, error) {
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
	if _, err := Client.Guild(s); err == nil {
		return s, nil
	}

	if _, err := Client.Channel(s); err == nil {
		return s, nil
	}

	return s, errInvalidID
}

func isBotAdmin(id string) bool {
	for _, admin := range Admins {
		if id == admin {
			return true
		}
	}
	return false
}

func NewMessage(m *dg.Message, msger core.Messenger) (*core.EventMessage, error) {
	a, err := NewAuthor(m.Author, m.Member, m.GuildID)
	if err != nil {
		return nil, err
	}

	h := &Here{
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		Author:    a,
	}

	sp := &Speaker{
		Author: a,
		Here:   h,
		VC:     nil,
	}

	return core.NewEventMessage(m.ID, m.Content, Frontend, a, h, msger, sp), nil
}

func isAdmin(guildID string, userID string) (bool, error) {
	return Client.MemberAllowed(guildID, userID, dg.PermissionAdministrator)
}

func isMod(guildID string, userID string) (bool, error) {
	return Client.MemberAllowed(guildID, userID, dg.PermissionBanMembers)
}

func msgSend(m *dg.Message, text string, embed *dg.MessageEmbed, ping bool) (*dg.Message, error) {
	// TODO: Consider adding an option which allows one of these 3 values
	// - no reply + no ping, just an embed
	// - reply + no ping (default)
	// - reply + ping
	// Maybe even no embed and just plain text?

	var embeds []*dg.MessageEmbed
	if embed != nil {
		embeds = append(embeds, embed)
	}

	var ref *dg.MessageReference
	// if there is no message id then sending a reference will return an error,
	// so instead we mention the user manually
	if m.ID == "" {
		text = m.Author.Mention() + " " + text
	} else {
		ref = m.Reference()
	}

	var mentions *dg.MessageAllowedMentions
	if !ping {
		mentions = &dg.MessageAllowedMentions{
			Parse: []dg.AllowedMentionType{},
		}
	}

	reply := &dg.MessageSend{
		Content:         text,
		Embeds:          embeds,
		AllowedMentions: mentions,
		Reference:       ref,
	}

	resp, err := Client.Session.ChannelMessageSendComplex(m.ChannelID, reply)
	if err != nil {
		return nil, err
	}

	// the returned response leaves the guild id empty whether or not the
	// message was sent to a guild, this makes it very difficult for the scope
	// getter functions to know which scope to return (they check if the guild
	// id is empty to know whether or not a message came from a DM), so we
	// manually set the guild id here
	resp.GuildID = m.GuildID

	core.RDB.Set(ctx, rdbMessageReplyToKeyPrefix+m.ID, resp.ID, 0)

	return resp, nil
}

func sendText(m *dg.Message, text string, ping bool) (*core.EventMessage, error) {
	var resp *dg.Message
	var err error

	lenLim := 2000
	// TODO: grapheme clusters instead of plain len?
	lenCnt := func(s string) int { return len(s) }

	if lenLim > lenCnt(text) {
		resp, err = msgSend(m, text, nil, ping)
	} else {
		parts := core.Split(text, lenCnt, lenLim)
		for _, p := range parts {
			resp, err = msgSend(m, p, nil, ping)
		}
	}

	if err != nil {
		return nil, err
	}
	return NewMessage(resp, &Message{Message: resp})
}

func sendEmbed(m *dg.Message, embed *dg.MessageEmbed, urr error, ping bool) (*core.EventMessage, error) {
	// TODO: implement message scrolling
	embed = embedColor(embed, urr)
	resp, err := msgSend(m, "", embed, ping)
	if err != nil {
		return nil, err
	}
	return NewMessage(resp, &Message{Message: resp})
}

func msgEdit(m *dg.Message, id, text string, embed *dg.MessageEmbed) (*dg.Message, error) {
	// Not using a var declaration for embeds because of the following scenario:
	// The original message contains an embed, our edit sends a text edit.
	// If the var declaration were to be used, the original embed would remain,
	// with the text edit added on top.
	//goland:noinspection GoPreferNilSlice
	embeds := []*dg.MessageEmbed{}
	if embed != nil {
		embeds = append(embeds, embed)
	}

	reply := &dg.MessageEdit{
		ID:      id,
		Channel: m.ChannelID,

		Content: &text,
		Embeds:  embeds,
		AllowedMentions: &dg.MessageAllowedMentions{
			Parse: []dg.AllowedMentionType{}, // don't ping user
		},
	}

	resp, err := Client.Session.ChannelMessageEditComplex(reply)
	if err != nil {
		return nil, err
	}

	// the returned response leaves the guild id empty whether or not the
	// message was sent to a guild, this makes it very difficult for the scope
	// getter functions to know which scope to return (they check if the guild
	// id is empty to know whether or not a message came from a DM), so we
	// manually set the guild id here
	resp.GuildID = m.GuildID

	return resp, nil
}

func editText(m *dg.Message, id, text string) (*core.EventMessage, error) {
	resp, err := msgEdit(m, id, text, nil)
	if err != nil {
		return nil, err
	}
	return NewMessage(resp, &Message{Message: resp})
}

func editEmbed(m *dg.Message, embed *dg.MessageEmbed, urr error, id string) (*core.EventMessage, error) {
	embed = embedColor(embed, urr)
	resp, err := msgEdit(m, id, "", embed)
	if err != nil {
		return nil, err
	}
	return NewMessage(resp, &Message{Message: resp})
}

func embedColor(embed *dg.MessageEmbed, urr error) *dg.MessageEmbed {
	if embed.Color != 0 {
		return embed
	}

	// default value of EmbedColor is 0 so even if it's not been set
	// then everything should be ok
	if urr == nil {
		embed.Color = EmbedColor
	} else {
		embed.Color = EmbedErrColor
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
