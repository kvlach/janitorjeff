package test

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Command = &core.CommandStatic{
	Names: []string{
		"test",
		"alias",
	},
	Description: "Test command.",
	Run:         run,
}
