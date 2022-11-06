package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"nick",
	},
	Frontends: frontends.All,
	Run:       adminRun,

	Children: core.Commands{
		{
			Names: core.Show,
			Run:   adminRunShow,
		},
		{
			Names: []string{
				"set",
			},
			UsageArgs: "<nick>",
			Run:       adminRunSet,
		},
		{
			Names:     core.Delete,
			UsageArgs: "<nick>",
			Run:       adminRunDelete,
		},
	},
}

func adminGetFlags(m *core.Message) (*flags, []string, error) {
	f := newFlags(m).Place().Person()
	args, err := f.fs.Parse()
	return f, args, err
}

func adminRun(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

//////////
//      //
// show //
//      //
//////////

func adminRunShow(m *core.Message) (any, error, error) {
	nick, usrErr, err := adminRunShowCore(m)
	if err != nil {
		return "", nil, err
	}
	return adminRunShowErr(usrErr, nick), usrErr, nil
}

func adminRunShowErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return nick
	case errPersonNotFound:
		return "nickname not set"
	default:
		return fmt.Sprint(usrErr)
	}
}

func adminRunShowCore(m *core.Message) (string, error, error) {
	fs, _, err := adminGetFlags(m)
	if err != nil {
		return "", nil, err
	}
	return runShow(fs.person, fs.place)
}

/////////
//     //
// set //
//     //
/////////

func adminRunSet(m *core.Message) (any, error, error) {
	_, usrErr, err := adminRunSetCore(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr == core.ErrMissingArgs {
		return m.Usage(), core.ErrMissingArgs, nil
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
	usrErr, err := runSet(nick, fs.person, fs.place)
	return nick, usrErr, err
}

////////////
//        //
// delete //
//        //
////////////

func adminRunDelete(m *core.Message) (any, error, error) {
	usrErr, err := adminRunDeleteCore(m)
	if err != nil {
		return "", nil, err
	}
	return adminRunDeleteErr(usrErr), usrErr, nil
}

func adminRunDeleteErr(usrErr error) string {
	switch usrErr {
	case nil:
		return "removed nick"
	case errPersonNotFound:
		return "person doesn't have a nickname in specified place"
	default:
		return fmt.Sprint(usrErr)
	}
}

func adminRunDeleteCore(m *core.Message) (error, error) {
	fs, _, err := adminGetFlags(m)
	if err != nil {
		return nil, err
	}
	return runDelete(fs.person, fs.place)
}
