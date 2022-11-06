package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Advanced = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Frontends:   frontends.All,
	Run:         advancedRun,
}

func advancedRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Advanced, m)
}
