package nick

import (
	"errors"
)

var (
	errUserNotFound = errors.New("user nick not found")
	errNickExists   = errors.New("nick is used by a different user")
)

func runGet(person, place int64) (string, error, error) {
	exists, err := dbUserExists(person, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errUserNotFound, nil
	}

	nick, err := dbUserNick(person, place)
	return nick, nil, err
}

func runSet(nick string, person, place int64) (error, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return nil, err
	}
	if nickExists {
		return errNickExists, nil
	}

	personExists, err := dbUserExists(person, place)
	if err != nil {
		return nil, err
	}

	if personExists {
		return nil, dbUserUpdate(person, place, nick)
	}
	return nil, dbUserAdd(person, place, nick)
}

func runDelete(person, place int64) (error, error) {
	exists, err := dbUserExists(person, place)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errUserNotFound, nil
	}
	return nil, dbUserDelete(person, place)
}
