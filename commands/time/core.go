package time

import (
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	"github.com/rs/zerolog/log"
	"github.com/tj/go-naturaldate"
)

var (
	errTimestamp        = errors.New("invalid timestamp")
	errTimezone         = errors.New("invalid timezone")
	errTimezoneNotSet   = errors.New("user hasn't set their timezone")
	errInvalidTime      = errors.New("could not parse given time string")
	errNoReminders      = errors.New("couldn't find any reminders")
	errReminderNotFound = errors.New("couldn't find person's reminder")
	errOldTime          = errors.New("given time has already passed")
)

type reminder struct {
	// Fields are public so that they show up in the debug logs
	ID     int64
	Person int64
	Place  int64
	When   time.Time
	What   string
	MsgID  string
}

//////////////
//          //
// database //
//          //
//////////////

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
);

CREATE TABLE IF NOT EXISTS CommandTimeReminders (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	person INTEGER NOT NULL,
	place INTEGER NOT NULL,
	time INTEGER NOT NULL,
	what VARCHAR(255) NOT NULL,
	msg_id VARCHAR(255) NOT NULL,
	FOREIGN KEY (person) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
);
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
	db.Lock.RLock()
	defer db.Lock.RUnlock()

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
	db.Lock.RLock()
	defer db.Lock.RUnlock()

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

func dbRemindAdd(person, place, when int64, what, msgID string) (int64, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	res, err := db.DB.Exec(`
	INSERT INTO CommandTimeReminders(person, place, time, what, msg_id)
	VALUES (?, ?, ?, ?, ?)`, person, place, when, what, msgID)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Int64("when", when).
		Str("what", what).
		Str("msgID", msgID).
		Msg("added reminder")

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func scanReminders(rows *sql.Rows) ([]reminder, error) {
	var rs []reminder
	for rows.Next() {
		var id, person, place, timestamp int64
		var what, msgID string
		if err := rows.Scan(&id, &person, &place, &timestamp, &what, &msgID); err != nil {
			return nil, err
		}
		r := reminder{
			ID:     id,
			Person: person,
			Place:  place,
			When:   time.Unix(timestamp, 0).UTC(),
			What:   what,
			MsgID:  msgID,
		}
		rs = append(rs, r)
		log.Debug().Interface("reminder", r).Msg("found reminder")
	}
	return rs, nil
}

func dbRemindList(person, place int64) ([]reminder, error) {
	db := core.Globals.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT id, person, place, time, what, msg_id
		FROM CommandTimeReminders
		WHERE person = ? and place = ?
	`, person, place)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs, err := scanReminders(rows)
	if err != nil {
		return nil, err
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Int("#reminders", len(rs)).
		Msg("got reminders")

	return rs, err
}

func dbRemindUpcoming(nowSeconds int64) ([]reminder, error) {
	db := core.Globals.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT id, person, place, time, what, msg_id
		FROM CommandTimeReminders
		WHERE time - ? < 300
	`, nowSeconds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rs, err := scanReminders(rows)
	if err != nil {
		return nil, err
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("when", nowSeconds).
		Int("#reminders", len(rs)).
		Msg("got upcoming reminders")

	return rs, err
}

func dbRemindDelete(id int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandTimeReminders
		WHERE id = ?`, id)

	log.Debug().
		Err(err).
		Int64("id", id).
		Msg("deleted reminder")

	return err
}

func dbRemindExists(id, person int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandTimeReminders
			WHERE id = ? and person = ?
			LIMIT 1
		)`, id, person)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("id", id).
		Int64("person", person).
		Bool("exists", exists).
		Msg("checked if reminder exists")

	return exists, err
}

/////////
//     //
// run //
//     //
/////////

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

func parseTime(when string, person, place int64) (time.Time, error, error) {
	tz, err := dbPersonTimezone(person, place)
	var loc *time.Location
	if err == nil {
		loc, err = time.LoadLocation(tz)
	} else {
		loc, err = time.LoadLocation("UTC")
	}

	if err != nil {
		return time.Time{}, nil, err
	}

	now := time.Now().In(loc)
	direction := naturaldate.WithDirection(naturaldate.Future)

	t, err := naturaldate.Parse(when, now, direction)
	if err != nil {
		return time.Time{}, errInvalidTime, nil
	}
	return t, nil, nil
}

func runTimestamp(when string, person, place int64) (time.Time, error, error) {
	t, usrErr, err := parseTime(when, person, place)
	if usrErr != nil || err != nil {
		return time.Time{}, usrErr, err
	}
	return t, nil, nil
}

func runTimezoneShow(person, place int64) (string, error, error) {
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

func runRemindAdd(when, what, msgID string, person, placeExact, placeLogical int64) (time.Time, int64, error, error) {
	t, usrErr, err := parseTime(when, person, placeLogical)
	if usrErr != nil || err != nil {
		return t, -1, usrErr, err
	}

	if t.Before(time.Now()) {
		return t, -1, errOldTime, nil
	}

	id, err := dbRemindAdd(person, placeExact, t.UTC().Unix(), what, msgID)

	// in case the reminder needs to happen close to immediately
	runUpcoming()

	return t, id, nil, err
}

func runRemindDelete(id, person int64) (error, error) {
	// if the person their own reminder, but from a different place then we
	// allow that
	exists, err := dbRemindExists(id, person)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errReminderNotFound, nil
	}
	return nil, dbRemindDelete(id)
}

func runRemindList(person, place int64) ([]reminder, error, error) {
	rs, err := dbRemindList(person, place)
	if err != nil {
		return nil, nil, err
	}
	if len(rs) == 0 {
		return nil, errNoReminders, nil
	}
	return rs, nil, nil
}

type upcoming struct {
	lock sync.RWMutex

	// serves as a set essentially
	waiting map[int64]struct{}
}

var upcomingReminders = upcoming{}

func (u *upcoming) add(r reminder) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.waiting == nil {
		u.waiting = map[int64]struct{}{}
	}

	if _, ok := u.waiting[r.ID]; ok {
		log.Debug().Int64("id", r.ID).Msg("reminder already in queue")
		return
	}

	u.waiting[r.ID] = struct{}{}
	log.Debug().Int64("id", r.ID).Msg("added reminder to queue")

	go func() {
		time.Sleep(r.When.Sub(time.Now()))

		m, err := frontends.CreateContext(r.Person, r.Place, r.MsgID)
		if err != nil {
			panic(err)
		}

		_, err = m.Client.Ping(r.What, nil)
		if err != nil {
			// TODO: retry
			panic(err)
		}

		err = dbRemindDelete(r.ID)
		if err != nil {
			// TODO
			panic(err)
		}
		u.del(r.ID)
	}()
}

func (u *upcoming) del(id int64) {
	u.lock.Lock()
	defer u.lock.Unlock()
	delete(u.waiting, id)
}

func runUpcoming() {
	rs, err := dbRemindUpcoming(time.Now().Unix())
	if err != nil {
		// TODO
		panic(err)
	}

	for _, r := range rs {
		upcomingReminders.add(r)
	}
}
