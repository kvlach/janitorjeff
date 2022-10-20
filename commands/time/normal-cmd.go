package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"time",
	},
	Description: "time stuff and things",
	UsageArgs:   "[<user> | zone]",
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