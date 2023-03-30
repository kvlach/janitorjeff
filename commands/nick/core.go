package nick

import (
	"errors"

	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

var (
	ErrPersonNotFound = errors.New("You have not set a nickname.")
	ErrNickExists     = errors.New("Nickname is already in use either by you or someone else.")
)

func dbNickExists(nick string, place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM settings_person
			WHERE cmd_nick_nick = $1 and place = $2
			LIMIT 1
		)`, nick, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Str("nick", nick).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if nick exists")

	return exists, err
}

func dbGetPerson(nick string, place int64) (int64, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var person int64

	row := db.DB.QueryRow(`
		SELECT person
		FROM settings_person
		WHERE cmd_nick_nick = $1 and place = $2`, nick, place)

	err := row.Scan(&person)

	log.Debug().
		Err(err).
		Str("nick", nick).
		Int64("place", place).
		Int64("person", person).
		Msg("got nick for person")

	return person, err
}

// Show returns the person's nickname in the specified place. If no nickname
// has been set then returns an ErrPersonNotFound error.
func Show(person, place int64) (string, error, error) {
	nick, err := core.DB.SettingPersonGet("cmd_nick_nick", person, place)
	if err != nil {
		return "", nil, err
	}
	if nick == nil {
		return "", ErrPersonNotFound, nil
	}
	return nick.(string), nil, nil
}

// Set sets the person's nickname in the specified place. If the nickname
// already exists in that place then it returns an ErrNickExists error.
func Set(nick string, person, place int64) (error, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return nil, err
	}
	if nickExists {
		return ErrNickExists, nil
	}
	return nil, core.DB.SettingPersonSet("cmd_nick_nick", person, place, nick)
}

// Delete deletes the person's nickname in the specified place. If no nickname
// has been set then returns an ErrPersonNotFound error.
func Delete(person, place int64) error {
	return core.DB.SettingPersonSet("cmd_nick_nick", person, place, nil)
}

// Tries to find a person from the given string. If "me" is passed the author
// is returned. Then tries to match a nickname and if it fails it tries various
// platform specific things (checking if the string is a mention of some sort,
// etc.)
func ParsePerson(m *core.Message, place int64, s string) (int64, error) {
	if s == "me" {
		return m.Author.Scope()
	}

	if person, err := dbGetPerson(s, place); err == nil {
		return person, nil
	}

	placeID, err := core.DB.ScopeID(place)
	if err != nil {
		return -1, err
	}

	id, err := m.Client.PersonID(s, placeID)
	if err != nil {
		return -1, err
	}

	return m.Client.Person(id)
}

// Same as ParsePerson but uses the default place instead
func ParsePersonHere(m *core.Message, s string) (int64, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return -1, err
	}
	return ParsePerson(m, here, s)
}
