package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Normal = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Frontends:   frontends.All,
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Normal, m)
}
