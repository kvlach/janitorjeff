package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

// Tries to find a person from the given string. First tries to match a nickname
// and if it fails it tries various platform specific things (checking if the
// string is a mention of some sort, etc.)
func ParsePerson(m *core.Message, place int64, s string) (int64, error) {
	if person, err := dbGetPerson(s, place); err == nil {
		return person, nil
	}

	placeID, err := core.Globals.DB.ScopeID(place)
	if err != nil {
		return -1, err
	}

	id, err := m.Client.PersonID(s, placeID)
	if err != nil {
		return -1, err
	}

	return m.Client.PersonScope(id)
}

// Same as ParsePerson but uses the default place instead
func ParsePersonHere(m *core.Message, s string) (int64, error) {
	here, err := m.ScopeHere()
	if err != nil {
		return -1, err
	}
	return ParsePerson(m, here, s)
}
