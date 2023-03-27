package discord

import (
	"database/sql"

	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS PlatformDiscordGuilds (
	id INTEGER PRIMARY KEY,
	guild VARCHAR(255) NOT NULL UNIQUE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS PlatformDiscordChannels (
	id INTEGER PRIMARY KEY,
	channel VARCHAR(255) NOT NULL UNIQUE,
	guild INTEGER NOT NULL,
	FOREIGN KEY (guild) REFERENCES PlatformDiscordGuilds(id) ON DELETE CASCADE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS PlatformDiscordUsers (
	id INTEGER PRIMARY KEY,
	uid VARCHAR(255) NOT NULL UNIQUE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);
`

func dbInit() error {
	return core.DB.Init(dbSchema)
}

func dbAddGuildScope(tx *sql.Tx, guildID string) (int64, error) {
	scope, err := core.DB.ScopeAdd(tx, guildID, Type)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO PlatformDiscordGuilds (id, guild)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;`, scope, guildID)

	if err != nil {
		return -1, err
	}

	return scope, nil
}

func dbAddChannelScope(tx *sql.Tx, channelID string, guildScope int64) (int64, error) {
	scope, err := core.DB.ScopeAdd(tx, channelID, Type)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO PlatformDiscordChannels(id, channel, guild)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;`, scope, channelID, guildScope)

	if err != nil {
		return -1, err
	}

	return scope, nil
}

func dbGetGuildScope(guildID string) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordGuilds
		WHERE guild = $1`, guildID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func dbGetChannelScope(channelID string) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordChannels
		WHERE channel = $1`, channelID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func dbGetGuildFromChannel(channelScope int64) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT guild
		FROM PlatformDiscordChannels
		WHERE id = $1`, channelScope)

	var guildScope int64
	err := row.Scan(&guildScope)
	return guildScope, err
}

func dbAddUserScope(tx *sql.Tx, userID string) (int64, error) {
	scope, err := core.DB.ScopeAdd(tx, userID, Type)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO PlatformDiscordUsers(id, uid)
		VALUES ($1, $2)`, scope, userID)

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Str("uid", userID).
		Msg("added user scope to db")

	return scope, err
}

func dbGetUserScope(userID string) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordUsers
		WHERE uid = $1`, userID)

	var id int64
	err := row.Scan(&id)

	log.Debug().
		Err(err).
		Int64("scope", id).
		Str("uid", userID).
		Msg("got user scope from db")

	return id, err
}
