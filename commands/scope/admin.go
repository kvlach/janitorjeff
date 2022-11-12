package scope

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

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
		"scope",
	}
}

func (admin) Description() string {
	return "Scope related commands."
}

func (admin) UsageArgs() string {
	return "(place | person)"
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminPlace,
		AdminPerson,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

///////////
//       //
// place //
//       //
///////////

var AdminPlace = adminPlace{}

type adminPlace struct{}

func (c adminPlace) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminPlace) Frontends() int {
	return c.Parent().Frontends()
}

func (adminPlace) Names() []string {
	return []string{
		"place",
	}
}

func (adminPlace) Description() string {
	return "Find a place's scope."
}

func (adminPlace) UsageArgs() string {
	return "<id>"
}

func (adminPlace) Parent() core.CommandStatic {
	return Admin
}

func (adminPlace) Children() core.CommandsStatic {
	return nil
}

func (adminPlace) Init() error {
	return nil
}

func (c adminPlace) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}
	place, err := c.core(m)
	return fmt.Sprint(place), nil, err
}

func (adminPlace) core(m *core.Message) (int64, error) {
	target := m.Command.Args[0]
	return runPlace(target, m.Client)
}

////////////
//        //
// person //
//        //
////////////

var AdminPerson = adminPerson{}

type adminPerson struct{}

func (c adminPerson) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminPerson) Frontends() int {
	return c.Parent().Frontends()
}

func (adminPerson) Names() []string {
	return []string{
		"person",
	}
}

func (adminPerson) Description() string {
	return "Find a person's scope."
}

func (adminPerson) UsageArgs() string {
	return "<id> [parent]"
}

func (adminPerson) Parent() core.CommandStatic {
	return Admin
}

func (adminPerson) Children() core.CommandsStatic {
	return nil
}

func (adminPerson) Init() error {
	return nil
}

func (c adminPerson) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.parent(m)
	case frontends.Twitch:
		return c.noParent(m)
	default:
		return "This command doesn't currently support this frontend.", nil, nil
	}
}

func (adminPerson) parent(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	target := m.Command.Args[0]
	parent := m.Command.Args[1]

	person, err := runPerson(target, parent, m.Client)
	return fmt.Sprint(person), nil, err
}

func (adminPerson) noParent(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	target := m.Command.Args[0]

	person, err := runPerson(target, "", m.Client)
	return fmt.Sprint(person), nil, err
}
