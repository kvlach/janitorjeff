package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandTimePeople (
	user INTEGER NOT NULL UNIQUE,
	timezone VARCHAR(255) NOT NULL,
	FOREIGN KEY (user) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func dbUserAdd(user int64, timezone string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	INSERT INTO CommandTimePeople(user, timezone)
	VALUES (?, ?)`, user, timezone)

	log.Debug().
		Err(err).
		Int64("user", user).
		Str("timezone", timezone).
		Msg("added user timezone in db")

	return err
}

func dbUserExists(user int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTimePeople
			WHERE user = ?
			LIMIT 1
		)`, user)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("user", user).
		Bool("exists", exists).
		Msg("checked db to see if scope exists")

	return exists, err

}

func dbUserDelete(user int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandTimePeople
		WHERE user = ?`, user)

	log.Debug().
		Err(err).
		Int64("user", user).
		Msg("deleted user from db")

	return err
}

func dbUserTimezone(user int64) (string, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var tz string

	row := db.DB.QueryRow(`
		SELECT timezone
		FROM CommandTimePeople
		WHERE user = ?`, user)

	err := row.Scan(&tz)

	log.Debug().
		Err(err).
		Int64("user", user).
		Str("timezone", tz).
		Msg("checked db to see if scope exists")

	return tz, err

}
