package nick

import (
	"errors"
)

var (
	errPersonNotFound = errors.New("user nick not found")
	errNickExists     = errors.New("nick is used by a different user")
)

// Show returns the person's nickname in the specified place. If no nickname
// has been set then returns an errPersonNotFound error.
func Show(person, place int64) (string, error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errPersonNotFound, nil
	}

	nick, err := dbPersonNick(person, place)
	return nick, nil, err
}

// Set sets the person's nickname in the specified place. If the person has set
// their nickname already then it updates it. If the nickname already exists in
// that place then it returns an errNickExists error.
func Set(nick string, person, place int64) (error, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return nil, err
	}
	if nickExists {
		return errNickExists, nil
	}

	personExists, err := dbPersonExists(person, place)
	if err != nil {
		return nil, err
	}

	if personExists {
		return nil, dbPersonUpdate(person, place, nick)
	}
	return nil, dbPersonAdd(person, place, nick)
}

// Delete deletes the person's nickname in the specified place. If no nickname
// has been set then returns an errPersonNotFound error.
func Delete(person, place int64) (error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errPersonNotFound, nil
	}
	return nil, dbPersonDelete(person, place)
}
