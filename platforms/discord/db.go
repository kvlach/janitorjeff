package discord

import (
	"database/sql"
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func getScope(type_ int, channelID, guildID string) (int64, error) {
	// In some cases a guild does not exist, for example in a DM, thus we are
	// forced to use the channel scope. A guild id is also not included in the
	// message object returned after sending a message, only the channel id is.
	// So, it can be difficult to differentiate if a message comes from a DM
	// or from a returned message object, since in order to do so we rely on
	// checking if the guild id field is empty. A way to solve this is relying
	// on the fact that the returned message object only comes after a user
	// has executed a command in that scope. That means that *if* a guild
	// exists it will already have been added in the database, and so we use
	// that.
	//
	// This can break if for example the bot were to send a message in a scope
	// where no message has ever been sent which means that channel/guild ids
	// have not been recorded, since the message create hook ignores the bot's
	// messages. This means that with the current implementation if that
	// were to happen in a guild, then channel scoped would be returned instead
	// of the guild scope. This is not a problem in places where no guild
	// exists.

	switch type_ {
	case Default, Guild, Channel, Thread:
		break
	default:
		return -1, fmt.Errorf("type '%d' not supported", type_)
	}

	if type_ == Thread {
		return -1, fmt.Errorf("Thread scopes not supported yet")
	}

	// if scope exists return it instead of re-adding it
	channelScope, err := getChannelScope(channelID)
	if err == nil {
		if type_ == Channel {
			return channelScope, nil
		}

		// find channel's guild scope
		// if channel exists then guild does also, even if it's the special
		// empty guild
		guildScope, err := getGuildFromChannel(channelScope)
		// A guild does exist even if it's a DM, it's the empty string guild,
		// so this means if there's an error, it's a different kind.
		if err != nil {
			return -1, err
		}

		if type_ == Guild {
			return guildScope, nil
		}

		// In the schema we make it so that the empty guild is the first one
		// added, and has `scope = 1`. This is the only way I can come up with
		// that doesn't require reading the DB again to search for the guild's
		// id
		if guildScope == 1 {
			return channelScope, err
		}
		return guildScope, nil
	}

	db := core.Globals.DB

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	// only create a new guildScope if it doesn't already exist
	guildScope, err := getGuildScope(guildID)
	if err != nil {
		guildScope, err = addGuildScope(tx, guildID)
		if err != nil {
			return -1, err
		}
	}

	channelScope, err = addChannelScope(tx, channelID, guildScope)
	if err != nil {
		return -1, err
	}

	var scope int64

	switch type_ {
	case Guild:
		scope = guildScope
	case Channel:
		scope = channelScope
	default:
		// We are sure that no guild exists here, which is why the channel is
		// returned
		if guildID == "" {
			scope = channelScope
		} else {
			scope = guildScope
		}
	}

	return scope, tx.Commit()
}

func addGuildScope(tx *sql.Tx, guildID string) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx)
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

func addChannelScope(tx *sql.Tx, channelID string, guildScope int64) (int64, error) {
	db := core.Globals.DB

	scope, err := db.ScopeAdd(tx)
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

func getGuildScope(guildID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordGuilds
		WHERE guild = ?`, guildID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func getChannelScope(channelID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformDiscordChannels
		WHERE channel = ?`, channelID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func getGuildFromChannel(channelScope int64) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT guild
		FROM PlatformDiscordChannels
		WHERE id = ?`, channelScope)

	var guildScope int64
	err := row.Scan(&guildScope)
	return guildScope, err
}
