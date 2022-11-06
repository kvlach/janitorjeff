package scope

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"scope",
	},
	Description: "Scope related commands.",
	UsageArgs:   "(place | person)",
	Frontends:   frontends.All,
	Run:         adminRun,

	Children: core.Commands{
		{
			Names: []string{
				"place",
			},
			Description: "Find a places's scope.",
			UsageArgs:   "<id>",
			Run:         adminRunPlace,
		},
		{
			Names: []string{
				"person",
			},
			Description: "Find a person's scope.",
			UsageArgs:   "<id> [parent]",
			Run:         adminRunPerson,
		},
	},
}

func adminRun(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

///////////
//       //
// place //
//       //
///////////

func adminRunPlace(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}
	place, err := adminRunPlaceCore(m)
	return fmt.Sprint(place), nil, err
}

func adminRunPlaceCore(m *core.Message) (int64, error) {
	target := m.Command.Runtime.Args[0]
	return runPlace(target, m.Client)
}

////////////
//        //
// person //
//        //
////////////

func adminRunPerson(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return adminRunPersonParent(m)
	case frontends.Twitch:
		return adminRunPersonNoParent(m)
	default:
		return "This command doesn't currently support this frontend.", nil, nil
	}
}

func adminRunPersonParent(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	target := m.Command.Runtime.Args[0]
	parent := m.Command.Runtime.Args[1]

	person, err := runPerson(target, parent, m.Client)
	return fmt.Sprint(person), nil, err
}

func adminRunPersonNoParent(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	target := m.Command.Runtime.Args[0]

	person, err := runPerson(target, "", m.Client)
	return fmt.Sprint(person), nil, err
}
