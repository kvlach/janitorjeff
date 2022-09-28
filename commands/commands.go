package commands

import (
	"git.slowtyper.com/slowtyper/janitorjeff/commands/help"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/prefix"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Commands = core.Commands{
	help.Command,
	prefix.Command,
}
