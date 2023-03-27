package prefix

import (
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

func dbAdd(prefix string, t core.CommandType, place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		INSERT INTO CommandPrefixPrefixes(place, prefix, type)
		VALUES ($1, $2, $3)`, place, prefix, t)

	log.Debug().
		Err(err).
		Str("prefix", prefix).
		Int("type", int(t)).
		Int64("place", place).
		Msg("added prefix")

	return err
}

func dbDelete(prefix string, place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandPrefixPrefixes
		WHERE prefix = $1 and place = $2`, prefix, place)

	log.Debug().
		Err(err).
		Str("prefix", prefix).
		Int64("place", place).
		Msg("deleted prefix")

	return err
}

func dbReset(place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandPrefixPrefixes
		WHERE place = $1`, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Msg("deleted all prefixes")

	return err
}
