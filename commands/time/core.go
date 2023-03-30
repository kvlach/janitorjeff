package time

import (
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/janitorjeff/jeff-bot/core"

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

func dbRemindAdd(person, place, when int64, what, msgID string) (int64, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var id int64
	err := db.DB.QueryRow(`
	INSERT INTO cmd_time_reminders(person, place, time, what, msg_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;`, person, place, when, what, msgID).Scan(&id)

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
	return id, nil
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
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT id, person, place, time, what, msg_id
		FROM cmd_time_reminders
		WHERE person = $1 and place = $2
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
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT id, person, place, time, what, msg_id
		FROM cmd_time_reminders
		WHERE time - $1 < 300
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
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM cmd_time_reminders
		WHERE id = $1`, id)

	log.Debug().
		Err(err).
		Int64("id", id).
		Msg("deleted reminder")

	return err
}

func dbRemindExists(id, person int64) (bool, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM cmd_time_reminders
			WHERE id = $1 and person = $2
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

func Now(person, place int64) (time.Time, error, error) {
	now := time.Now().UTC()

	tz, err := core.DB.SettingPersonGet("cmd_time_tz", person, place)
	if err != nil {
		return now, nil, err
	}

	loc, err := time.LoadLocation(tz.(string))
	if err != nil {
		return now, nil, err
	}

	return now.In(loc), nil, nil
}

func Convert(target, tz string) (string, error, error) {
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

func Time(when string, person, place int64) (time.Time, error, error) {
	tz, err := core.DB.SettingPersonGet("cmd_time_tz", person, place)
	loc, err := time.LoadLocation(tz.(string))
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

func Timestamp(when string, person, place int64) (time.Time, error, error) {
	t, usrErr, err := Time(when, person, place)
	if usrErr != nil || err != nil {
		return time.Time{}, usrErr, err
	}
	return t, nil, nil
}

func TimezoneShow(person, place int64) (string, error) {
	tz, err := core.DB.SettingPersonGet("cmd_time_tz", person, place)
	if err != nil {
		return "", err
	}
	return tz.(string), nil
}

func TimezoneSet(tz string, person, place int64) (string, error, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return tz, errTimezone, nil
	}
	tz = loc.String()
	return tz, nil, core.DB.SettingPersonSet("cmd_time_tz", person, place, tz)
}

func TimezoneDelete(person, place int64) error {
	return core.DB.SettingPersonSet("cmd_time_tz", person, place, "UTC")
}

func RemindAdd(when, what, msgID string, person, placeExact, placeLogical int64) (time.Time, int64, error, error) {
	t, usrErr, err := Time(when, person, placeLogical)
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

func RemindDelete(id, person int64) (error, error) {
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

func RemindList(person, place int64) ([]reminder, error, error) {
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

		m, err := core.Frontends.CreateMessage(r.Person, r.Place, r.MsgID)
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
