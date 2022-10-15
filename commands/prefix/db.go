package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

func dbScopeExists(scope int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandPrefixPrefixes
			WHERE scope = ?
			LIMIT 1
		)`, scope)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Bool("exists", exists).
		Msg("checked db to see if scope exists")

	return exists, err
}

func dbPrefixExists(prefix string, scope int64, t int) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandPrefixPrefixes
			WHERE scope = ? and prefix = ? and type = ?
			LIMIT 1
		)`, scope, prefix, t)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Str("prefix", prefix).
		Int64("scope", scope).
		Bool("exists", exists).
		Msg("checked db to see if prefix exists")

	return exists, err
}

func dbAdd(prefix string, scope int64, t int) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		INSERT INTO CommandPrefixPrefixes(scope, prefix, type)
		VALUES (?, ?, ?)`, scope, prefix, t)

	log.Debug().
		Err(err).
		Str("prefix", prefix).
		Int64("scope", scope).
		Msg("added prefix")

	return err
}

func dbDel(prefix string, scope int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandPrefixPrefixes
		WHERE prefix = ? and scope = ?`, prefix, scope)

	log.Debug().
		Err(err).
		Str("prefix", prefix).
		Int64("scope", scope).
		Msg("deleted prefix")

	return err
}

func dbReset(scope int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandPrefixPrefixes
		WHERE scope = ?`, scope)

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Msg("deleted all prefixes")

	return err
}
