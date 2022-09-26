package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Command = &core.CommandStatic{
	Names: []string{
		"help",
	},
	Description: "Shows the help message of the specified command.",
	UsageArgs:   "<command>",
	Run:         run,
}
