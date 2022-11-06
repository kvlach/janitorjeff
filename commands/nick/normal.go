package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"nick",
		"nickname",
	},
	Description: "View or set your nickname.",
	UsageArgs:   "[nickname]",
	Frontends:   frontends.All,
	Run:         normalRun,
	Init:        init_,
}

func normalRun(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) == 0 {
		return advancedRunShow(m)
	}
	return advancedRunSet(m)
}
