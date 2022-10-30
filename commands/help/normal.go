package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Normal, m)
}
