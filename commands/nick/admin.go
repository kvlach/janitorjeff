package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"nick",
	},
	Run: adminRun,

	Children: core.Commands{
		{
			Names: []string{
				"get",
			},
			Run: adminRunGet,
		},
		{
			Names: []string{
				"set",
			},
			Run: adminRunSet,
		},
	},
}

func adminGetFlags(m *core.Message) (*flags, []string, error) {
	f := newFlags(m).Place().Person()
	args, err := f.fs.Parse()
	return f, args, err
}

func adminRun(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

func adminRunGet(m *core.Message) (any, error, error) {
	nick, usrErr, err := adminRunGetCore(m)
	if err != nil {
		return "", nil, err
	}
	return adminRunGetErr(usrErr, nick), usrErr, nil
}

func adminRunGetErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return nick
	case errUserNotFound:
		return "nickname not set"
	default:
		return fmt.Sprint(usrErr)
	}
}

func adminRunGetCore(m *core.Message) (string, error, error) {
	fs, _, err := adminGetFlags(m)
	if err != nil {
		return "", nil, err
	}
	return runGet(fs.person, fs.place)
}

func adminRunSet(m *core.Message) (any, error, error) {
	_, usrErr, err := adminRunSetCore(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr == core.ErrMissingArgs {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}
	return adminRunSetErr(usrErr), usrErr, nil
}

func adminRunSetErr(usrErr error) string {
	switch usrErr {
	case nil:
		return "set nickname"
	case errNickExists:
		return "nickname already exists"
	default:
		return fmt.Sprint(usrErr)
	}
}

func adminRunSetCore(m *core.Message) (string, error, error) {
	fs, args, err := adminGetFlags(m)
	if err != nil {
		return "", nil, err
	}
	if len(args) == 0 {
		return "", core.ErrMissingArgs, nil
	}
	nick := args[0]
	return runSet(nick, fs.person, fs.place)
}
