package tiktok

import (
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandTTSCustomUserVoices (
	person INTEGER NOT NULL,
	place INTEGER NOT NULL,
	voice VARCHAR(255) NOT NULL,
	UNIQUE(person, place)
);

-- All settings must come with default values, as those are used when first
-- adding a new entry.
CREATE TABLE IF NOT EXISTS CommandTTSPlaceSettings (
	place INTEGER NOT NULL UNIQUE,
	subonly BOOLEAN NOT NULL DEFAULT FALSE
);
`

func dbPersonVoiceExists(person, place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTTSCustomUserVoices
			WHERE person = ? and place = ?
			LIMIT 1
		)`, person, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if voice exists")

	return exists, err
}

func dbPersonGetVoice(person, place int64) (string, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var voice string

	row := db.DB.QueryRow(`
		SELECT voice
		FROM CommandTTSCustomUserVoices
		WHERE person = ? and place = ?`, person, place)

	err := row.Scan(&voice)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("voice", voice).
		Msg("got voice for person")

	return voice, err
}

func dbPersonAddVoice(person, place int64, voice string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		INSERT INTO CommandTTSCustomUserVoices(person, place, voice)
		VALUES (?, ?, ?)`, person, place, voice)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("voice", voice).
		Msg("added person voice in db")

	return err
}

func dbPersonUpdateVoice(person, place int64, voice string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE CommandTTSCustomUserVoices
		SET voice = ?
		WHERE person = ? and place = ?
	`, voice, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("voice", voice).
		Msg("updated person voice")

	return err
}

func dbPlaceSettingsExist(place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTTSPlaceSettings
			WHERE place = ?
			LIMIT 1
		)`, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if place settings exist")

	return exists, err
}

// dbPlaceSettingsGenerate will just use all of the settings' default values,
// as defined in the schema.
func dbPlaceSettingsGenerate(place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		INSERT INTO CommandTTSPlaceSettings(place)
		VALUES (?)`, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Msg("generated settings")

	return err
}

// Assumes that the settings for the specified place already exists.
func dbSubOnlyGet(place int64) (bool, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var subonly bool

	row := db.DB.QueryRow(`
		SELECT subonly
		FROM CommandTTSPlaceSettings
		WHERE place = ?`, place)

	err := row.Scan(&subonly)

	log.Debug().
		Err(err).
		Int64("place", place).
		Bool("subonly", subonly).
		Msg("got subonly setting")

	return subonly, err
}

// Assumes that the settings for the specified place already exists.
func dbSubOnlySet(place int64, subonly bool) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE CommandTTSPlaceSettings
		SET subonly = ?
		WHERE place = ?
	`, subonly, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Bool("subonly", subonly).
		Msg("updated subonly setting")

	return err
}
