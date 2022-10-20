package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"nick",
	},
	Run: runAdmin,

	Children: core.Commands{
		{
			Names: []string{
				"get",
			},
			Run: runAdminGet,
		},
		{
			Names: []string{
				"set",
			},
			Run: runAdminSet,
		},
	},
}

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
