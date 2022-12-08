package nick

import (
	"errors"

	"github.com/janitorjeff/jeff-bot/core"
)

var (
	errPersonNotFound = errors.New("user nick not found")
	errNickExists     = errors.New("nick is used by a different user")
)

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

//////////
//      //
// init //
//      //
//////////

func init_() error {
	return core.DB.Init(dbSchema)
}
