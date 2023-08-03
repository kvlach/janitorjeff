package streak

import (
	"fmt"
	"strconv"

	"git.sr.ht/~slowtyper/janitorjeff/commands/nick"
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(m *core.Message) bool {
	return m.Frontend.Type() == twitch.Type
}

func (admin) Names() []string {
	return Advanced.Names()
}

func (admin) Description() string {
	return Advanced.Description()
}

func (c admin) UsageArgs() string {
	return c.Children().Usage()
}

func (admin) Category() core.CommandCategory {
	return Advanced.Category()
}

func (admin) Examples() []string {
	return nil
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminShow,
		AdminSet,
	}

}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (resp any, urr core.Urr, err error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c adminShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminShow) Names() []string {
	return core.AliasesShow
}

func (adminShow) Description() string {
	return "Show the user's streak"
}

func (c adminShow) UsageArgs() string {
	return "<user>"
}

func (c adminShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminShow) Examples() []string {
	return nil
}

func (adminShow) Parent() core.CommandStatic {
	return Admin
}

func (adminShow) Children() core.CommandsStatic {
	return nil
}

func (adminShow) Init() error {
	return nil
}

func (adminShow) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}
	person, err := nick.ParsePersonHere(m, m.Command.Args[0])
	if err != nil {
		return nil, nil, err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, nil, err
	}
	streak, err := Get(person, here)
	if err != nil {
		return nil, nil, err
	}
	return fmt.Sprintf("Streak is: %d", streak), nil, nil
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

func (c adminSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminSet) Names() []string {
	return core.AliasesSet
}

func (adminSet) Description() string {
	return "Set the user's streak"
}

func (c adminSet) UsageArgs() string {
	return "<user> <streak>"
}

func (c adminSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminSet) Examples() []string {
	return nil
}

func (adminSet) Parent() core.CommandStatic {
	return Admin
}

func (adminSet) Children() core.CommandsStatic {
	return nil
}

func (adminSet) Init() error {
	return nil
}

func (adminSet) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.UrrMissingArgs, nil
	}
	person, err := nick.ParsePersonHere(m, m.Command.Args[0])
	if err != nil {
		return nil, nil, err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, nil, err
	}
	streak, err := strconv.Atoi(m.Command.Args[1])
	if err != nil {
		return "Expected number, got: " + m.Command.Args[1], core.UrrNew("not a number"), nil
	}
	if err := Set(person, here, streak); err != nil {
		return nil, nil, err
	}
	return fmt.Sprintf("Updated streak to %d.", streak), nil, nil
}
