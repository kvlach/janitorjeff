package nick

import (
	"github.com/kvlach/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var (
	UrrNickExists = core.UrrNew("Nickname is already in use either by you or someone else.")
)

func dbNickExists(nick string, place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM info_person
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
		FROM info_person
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

// Show returns the person's nickname in the specified place.
// If the nickname is nil returns core.UrrValNil.
func Show(person, place int64) (string, core.Urr, error) {
	return core.DB.PersonGet("cmd_nick_nick", person, place).StrNil()
}

// Set sets the person's nickname in the specified place.
// If the nickname is already in use in the place returns UrrNickExists.
func Set(nick string, person, place int64) (core.Urr, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return nil, err
	}
	if nickExists {
		return UrrNickExists, nil
	}
	return nil, core.DB.PersonSet("cmd_nick_nick", person, place, nick)
}

// Delete sets the person's nickname in the specified place to nil.
func Delete(person, place int64) error {
	return core.DB.PersonSet("cmd_nick_nick", person, place, nil)
}

// ParsePerson tries to find a person from the given string.
// If "me" is passed, the author is returned, otherwise tries to match a nickname,
// and if it fails, it tries various frontend-specific extraction methods
// (e.g., checking if the string is a mention of some sort, etc.)
func ParsePerson(m *core.EventMessage, place int64, s string) (int64, error) {
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

// ParsePersonHere is the same as ParsePerson but place = here.
func ParsePersonHere(m *core.EventMessage, s string) (int64, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return -1, err
	}
	return ParsePerson(m, here, s)
}
