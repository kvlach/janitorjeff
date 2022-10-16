package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"nick",
		"nickname",
	},
	Description: "Set a nickname that can be used when calling commands.",
	UsageArgs:   "<nickname>",
	Run:         runNormal,
	Init:        init_,
}
