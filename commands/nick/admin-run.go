package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func getAdminFlags(m *core.Message) (*flags, []string, error) {
	f := newFlags(m).Place().Person()
	args, err := f.fs.Parse()
	return f, args, err
}

func runAdmin(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

func runAdminGet(m *core.Message) (any, error, error) {
	nick, usrErr, err := runAdminGetCore(m)
	if err != nil {
		return "", nil, err
	}
	return runAdminGetErr(usrErr, nick), usrErr, nil
}

func runAdminGetErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return nick
	case errUserNotFound:
		return "nickname not set"
	default:
		return fmt.Sprint(usrErr)
	}
}

func runAdminGetCore(m *core.Message) (string, error, error) {
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return "", nil, err
	}
	return runGet(fs.person, fs.place)
}

func runAdminSet(m *core.Message) (any, error, error) {
	_, usrErr, err := runAdminSetCore(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr == core.ErrMissingArgs {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}
	return runAdminSetErr(usrErr), usrErr, nil
}

func runAdminSetErr(usrErr error) string {
	switch usrErr {
	case nil:
		return "set nickname"
	case errNickExists:
		return "nickname already exists"
	default:
		return fmt.Sprint(usrErr)
	}
}

func runAdminSetCore(m *core.Message) (string, error, error) {
	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", nil, err
	}
	if len(args) == 0 {
		return "", core.ErrMissingArgs, nil
	}
	nick := args[0]
	return runSet(nick, fs.person, fs.place)
}

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

func runSet(nick string, person, place int64) (string, error, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return "", nil, err
	}
	if nickExists {
		return nick, errNickExists, nil
	}

	personExists, err := dbUserExists(person, place)
	if err != nil {
		return "", nil, err
	}

	if personExists {
		return nick, nil, dbUserUpdate(person, place, nick)
	}
	return nick, nil, dbUserAdd(person, place, nick)
}
