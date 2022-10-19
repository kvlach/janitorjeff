package discord

import (
	"database/sql"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func getPlaceScope(id string, m *dg.Message, s *dg.Session) (int64, error) {
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
		return -1, err
	}

	db := core.Globals.DB

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
	guildScope, err := dbGetGuildScope(guildID)
	if err != nil {
		guildScope, err = dbAddGuildScope(tx, guildID)
		if err != nil {
			return -1, err
		}
	}

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

func dbAddGuildScope(tx *sql.Tx, guildID string) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx, guildID)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO PlatformDiscordGuilds(id, guild)
		VALUES (?, ?)`, scope, guildID)

	if err != nil {
		return -1, err
	}

	return scope, nil
}

func dbAddChannelScope(tx *sql.Tx, channelID string, guildScope int64) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx, channelID)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO PlatformDiscordChannels(id, channel, guild)
		VALUES (?, ?, ?)`, scope, channelID, guildScope)

	if err != nil {
		return -1, err
	}

	return scope, nil
}

func dbGetGuildScope(guildID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordGuilds
		WHERE guild = ?`, guildID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func dbGetChannelScope(channelID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordChannels
		WHERE channel = ?`, channelID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func dbGetGuildFromChannel(channelScope int64) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT guild
		FROM PlatformDiscordChannels
		WHERE id = ?`, channelScope)

	var guildScope int64
	err := row.Scan(&guildScope)
	return guildScope, err
}

func dbAddUserScope(tx *sql.Tx, userID string) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx, userID)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO PlatformDiscordUsers(id, user)
		VALUES (?, ?)`, scope, userID)

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Str("user", userID).
		Msg("added user scope to db")

	return scope, err
}

func dbGetUserScope(userID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordUsers
		WHERE user = ?`, userID)

	var id int64
	err := row.Scan(&id)

	log.Debug().
		Err(err).
		Int64("scope", id).
		Str("user", userID).
		Msg("got user scope from db")

	return id, err
}
