package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

func adminGetFlags(m *core.Message) (*flags, []string, error) {
	f := newFlags(m).Place().Person()
	args, err := f.fs.Parse()
	return f, args, err
}

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Frontends() int {
	return frontends.All
}

func (admin) Names() []string {
	return []string{
		"nick",
		"nickname",
	}
}

func (admin) Description() string {
	return ""
}

func (admin) UsageArgs() string {
	return ""
}

func (admin) Parent() core.Commander {
	return nil
}

func (admin) Children() core.Commanders {
	return core.Commanders{
		AdminShow,
		AdminSet,
		AdminDelete,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

//////////
//      //
// show //
//      //
//////////

var AdminShow = adminShow{}

type adminShow struct{}

func (c adminShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminShow) Frontends() int {
	return c.Parent().Frontends()
}

func (adminShow) Names() []string {
	return core.Show
}

func (adminShow) Description() string {
	return ""
}

func (adminShow) UsageArgs() string {
	return ""
}

func (adminShow) Parent() core.Commander {
	return Admin
}

func (adminShow) Children() core.Commanders {
	return nil
}

func (adminShow) Init() error {
	return nil
}

func (c adminShow) Run(m *core.Message) (any, error, error) {
	nick, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr, nick), usrErr, nil
}

func (adminShow) err(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return nick
	case errPersonNotFound:
		return "nickname not set"
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminShow) core(m *core.Message) (string, error, error) {
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

var AdminSet = adminSet{}

type adminSet struct{}

func (c adminSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminSet) Frontends() int {
	return c.Parent().Frontends()
}

func (adminSet) Names() []string {
	return []string{
		"set",
	}
}

func (adminSet) Description() string {
	return ""
}

func (adminSet) UsageArgs() string {
	return "<nick>"
}

func (adminSet) Parent() core.Commander {
	return Admin
}

func (adminSet) Children() core.Commanders {
	return nil
}

func (adminSet) Init() error {
	return nil
}

func (c adminSet) Run(m *core.Message) (any, error, error) {
	_, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr == core.ErrMissingArgs {
		return m.Usage(), core.ErrMissingArgs, nil
	}
	return c.err(usrErr), usrErr, nil
}

func (adminSet) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "set nickname"
	case errNickExists:
		return "nickname already exists"
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminSet) core(m *core.Message) (string, error, error) {
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

var AdminDelete = adminDelete{}

type adminDelete struct{}

func (c adminDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminDelete) Frontends() int {
	return c.Parent().Frontends()
}

func (adminDelete) Names() []string {
	return core.Delete
}

func (adminDelete) Description() string {
	return ""
}

func (adminDelete) UsageArgs() string {
	return "<nick>"
}

func (adminDelete) Parent() core.Commander {
	return Admin
}

func (adminDelete) Children() core.Commanders {
	return nil
}

func (adminDelete) Init() error {
	return nil
}

func (c adminDelete) Run(m *core.Message) (any, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (adminDelete) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "removed nick"
	case errPersonNotFound:
		return "person doesn't have a nickname in specified place"
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminDelete) core(m *core.Message) (error, error) {
	fs, _, err := adminGetFlags(m)
	if err != nil {
		return nil, err
	}
	return runDelete(fs.person, fs.place)
}
