package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Advanced = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Run:         advancedRun,
}

func advancedRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Advanced, m)
}
