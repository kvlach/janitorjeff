package commands

import (
	"git.slowtyper.com/slowtyper/janitorjeff/commands/command"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/help"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/prefix"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/test"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = core.Commands{
	command.Command,
	help.Command,
	prefix.Command,
}

var Advanced = core.Commands{}

var Admin = core.Commands{
	test.Command,
}
