package discord

import (
	"database/sql"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

func dbAddGuildScope(tx *sql.Tx, guildID string) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx, guildID, Type)
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

	scope, err := db.ScopeAdd(tx, channelID, Type)
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

	scope, err := db.ScopeAdd(tx, userID, Type)
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
