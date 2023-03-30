package discord

import (
	"database/sql"

	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

func dbAddGuildScope(tx *sql.Tx, guildID string) (int64, error) {
	scope, err := core.DB.ScopeAdd(tx, guildID, Type)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO frontend_discord_guilds (scope, guild)
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
		INSERT INTO frontend_discord_channels (scope, channel, guild)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;`, scope, channelID, guildScope)

	if err != nil {
		return -1, err
	}

	return scope, nil
}

func dbGetGuildScope(guildID string) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT scope
		FROM frontend_discord_guilds
		WHERE guild = $1`, guildID)

	var scope int64
	err := row.Scan(&scope)
	return scope, err
}

func dbGetChannelScope(channelID string) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT scope
		FROM frontend_discord_channels
		WHERE channel = $1`, channelID)

	var scope int64
	err := row.Scan(&scope)
	return scope, err
}

func dbGetGuildFromChannel(channelScope int64) (int64, error) {
	row := core.DB.DB.QueryRow(`
		SELECT guild
		FROM frontend_discord_channels
		WHERE scope = $1`, channelScope)

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
		INSERT INTO frontend_discord_users(scope, uid)
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
		SELECT scope
		FROM frontend_discord_users
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
