package test

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Command = &core.CommandStatic{
	Names: []string{
		"test",
		"alias",
	},
	Description: "Test command.",
	Frontends:   frontends.All,
	Run:         run,
}
