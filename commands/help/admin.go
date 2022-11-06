package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Admin = &core.CommandStatic{
	Names:       cmdAliases,
	Description: cmdDescription,
	UsageArgs:   cmdUsageArgs,
	Frontends:   frontends.All,
	Run:         adminRun,
}

func adminRun(m *core.Message) (any, error, error) {
	return run(core.Globals.Commands.Admin, m)
}
