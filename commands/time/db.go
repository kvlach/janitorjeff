package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

// A place is specified for timezones, even though it's not necessary (someone's
// timezone isn't going to change if they call the command from a different
// place), in order to avoid situations in which someone's approximate location
// can be given away just by checking the timezone they set in a different
// place.

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandTimePeople (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user INTEGER NOT NULL,
	place INTEGER NOT NULL ,
	timezone VARCHAR(255) NOT NULL,
	UNIQUE(user, place),
	FOREIGN KEY (user) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func dbUserAdd(user, place int64, timezone string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	INSERT INTO CommandTimePeople(user, place, timezone)
	VALUES (?, ?, ?)`, user, place, timezone)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Str("timezone", timezone).
		Msg("added user timezone in db")

	return err
}

func dbUserExists(user, place int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTimePeople
			WHERE user = ? and place = ?
			LIMIT 1
		)`, user, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if user exists")

	return exists, err

}

func dbUserDelete(user, place int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandTimePeople
		WHERE user = ? and place = ?`, user, place)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Msg("deleted user from db")

	return err
}

func dbUserTimezone(user, place int64) (string, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var tz string

	row := db.DB.QueryRow(`
		SELECT timezone
		FROM CommandTimePeople
		WHERE user = ? and place = ?`, user, place)

	err := row.Scan(&tz)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Str("timezone", tz).
		Msg("got timezone from db")

	return tz, err

}
