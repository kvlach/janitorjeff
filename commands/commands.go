package commands

import (
	"git.slowtyper.com/slowtyper/janitorjeff/commands/command"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/connect"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/help"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/nick"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/paintball"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/prefix"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/rps"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/scope"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/test"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/time"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/urban-dictionary"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/wikipedia"
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var Normal = core.Commands{
	command.Command,
	connect.Normal,
	help.Normal,
	nick.Normal,
	paintball.Normal,
	prefix.Command,
	rps.Normal,
	time.Normal,
	urban_dictionary.Normal,
	wikipedia.Normal,
}

var Advanced = core.Commands{
	help.Advanced,
	nick.Advanced,
	time.Advanced,
	urban_dictionary.Advanced,
}

var Admin = core.Commands{
	help.Admin,
	nick.Admin,
	prefix.Admin,
	scope.Admin,
	test.Command,
}

var All = core.AllCommands{
	Normal:   Normal,
	Advanced: Advanced,
	Admin:    Admin,
}

func setup(cmd *core.CommandStatic) {
	if cmd.Frontends == 0 {
		panic("frontends not set for command: " + cmd.Format(""))
	}

	if cmd.Children == nil {
		return
	}

	for _, child := range cmd.Children {
		// Setting the parents when declaring the object is not possible because
		// that results in an inialization loop error (children reference the
		// parent, so the parent can't reference the children). This also makes
		// declaring the commands a bit cleaner.
		child.Parent = cmd

		// child inherits parent's Frontend
		child.Frontends = cmd.Frontends

		setup(child)
	}
}

func init() {
	for _, cmd := range Normal {
		setup(cmd)
	}

	for _, cmd := range Advanced {
		setup(cmd)
	}

	for _, cmd := range Admin {
		setup(cmd)
	}
}

func runInit(cmd *core.CommandStatic) {
	if cmd.Init != nil {
		if err := cmd.Init(); err != nil {
			log.Fatal().Err(err).Msgf("failed to init command %v", cmd)
		}
	}
}

// This must be run after all of the global variables have been set (including
// ones that frontend init functions might set) since the `Init` functions might
// depend on them.
func Init() {
	for _, cmd := range Normal {
		runInit(cmd)
	}

	for _, cmd := range Advanced {
		runInit(cmd)
	}

	for _, cmd := range Admin {
		runInit(cmd)
	}
}
