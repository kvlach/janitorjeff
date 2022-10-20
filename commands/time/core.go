package time

import (
	"errors"
	"strconv"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var (
	errTimestamp      = errors.New("invalid timestamp")
	errTimezone       = errors.New("invalid timezone")
	errTimezoneNotSet = errors.New("user hasn't set their timezone")
)

// A place is specified for timezones, even though it's not necessary (someone's
// timezone isn't going to change if they call the command from a different
// place), in order to avoid situations in which someone's approximate location
// can be given away just by checking the timezone they set in a different
// place.

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandTimePeople (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	person INTEGER NOT NULL,
	place INTEGER NOT NULL ,
	timezone VARCHAR(255) NOT NULL,
	UNIQUE(person, place),
	FOREIGN KEY (person) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func dbPersonAdd(person, place int64, timezone string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	INSERT INTO CommandTimePeople(person, place, timezone)
	VALUES (?, ?, ?)`, person, place, timezone)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("timezone", timezone).
		Msg("added person timezone in db")

	return err
}

func dbPersonUpdate(person, place int64, timezone string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	UPDATE CommandTimePeople
	SET timezone = ?
	WHERE person = ? and place = ?
	`, timezone, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("timezone", timezone).
		Msg("updated timezone")

	return err
}

func dbPersonExists(person, place int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTimePeople
			WHERE person = ? and place = ?
			LIMIT 1
		)`, person, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if person exists")

	return exists, err

}

func dbPersonDelete(person, place int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandTimePeople
		WHERE person = ? and place = ?`, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Msg("deleted person from db")

	return err
}

func dbPersonTimezone(person, place int64) (string, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var tz string

	row := db.DB.QueryRow(`
		SELECT timezone
		FROM CommandTimePeople
		WHERE person = ? and place = ?`, person, place)

	err := row.Scan(&tz)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("timezone", tz).
		Msg("got timezone from db")

	return tz, err

}

func runNow(person, place int64) (time.Time, error, error) {
	now := time.Now().UTC()

	exists, err := dbPersonExists(person, place)
	if err != nil {
		return now, nil, err
	}

	if !exists {
		return now, errTimezoneNotSet, nil
	}

	tz, err := dbPersonTimezone(person, place)
	if err != nil {
		return now, nil, err
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return now, nil, err
	}

	return now.In(loc), nil, nil
}

func runConvert(target, tz string) (string, error, error) {
	var t time.Time
	if target == "now" {
		t = time.Now().UTC()
	} else {
		timestamp, err := strconv.ParseInt(target, 10, 64)
		if err != nil {
			return "", errTimestamp, nil
		}
		t = time.Unix(timestamp, 0).UTC()
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", errTimezone, nil
	}

	return t.In(loc).Format(time.UnixDate), nil, nil
}

func runTimezoneGet(person, place int64) (string, error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errTimezoneNotSet, nil
	}

	tz, err := dbPersonTimezone(person, place)
	if err != nil {
		return "", nil, err
	}

	return tz, nil, nil
}

func runTimezoneSet(tz string, person, place int64) (string, error, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return tz, errTimezone, nil
	}

	tz = loc.String()

	exists, err := dbPersonExists(person, place)
	if err != nil {
		return tz, nil, err
	}

	if exists {
		return tz, nil, dbPersonUpdate(person, place, tz)
	}
	return tz, nil, dbPersonAdd(person, place, tz)
}

func runTimezoneDelete(person, place int64) (error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errTimezoneNotSet, nil
	}
	return nil, dbPersonDelete(person, place)
}
