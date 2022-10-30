package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Admin = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Run:         adminRun,
}

func adminRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Admin, m)
}
