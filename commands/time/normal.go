package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"time",
	},
	Description: "time stuff and things",
	UsageArgs:   "[<user> | zone]",
	Frontends:   frontends.All,
	Run:         runNormal,
	Init:        init_,

	Children: core.Commands{
		cmdNormalTimezone,
	},
}

var cmdNormalTimezone = &core.CommandStatic{
	Names: []string{
		"zone",
	},
	Description: "Set or view your own timezone.",
	UsageArgs:   "[timezone]",
	Run:         runNormalTimezone,
}

func runNormal(m *core.Message) (any, error, error) {
	return advancedRunNow(m)
}

func runNormalTimezone(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) == 0 {
		return advancedRunTimezoneShow(m)
	}
	return advancedRunTimezoneSet(m)
}
