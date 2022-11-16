package commands

import (
	"git.slowtyper.com/slowtyper/janitorjeff/commands/category"
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
	"git.slowtyper.com/slowtyper/janitorjeff/commands/title"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/urban-dictionary"
	"git.slowtyper.com/slowtyper/janitorjeff/commands/wikipedia"
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var Commands = core.CommandsStatic{
	category.Normal,

	command.Normal,

	connect.Normal,

	help.Normal,
	help.Advanced,
	help.Admin,

	nick.Normal,
	nick.Advanced,
	nick.Admin,

	paintball.Normal,

	prefix.Normal,
	prefix.Admin,

	rps.Normal,

	scope.Admin,

	test.Admin,

	time.Normal,
	time.Advanced,

	title.Normal,
	title.Advanced,

	urban_dictionary.Normal,
	urban_dictionary.Advanced,

	wikipedia.Normal,
}

// This must be run after all of the global variables have been set (including
// ones that frontend init functions might set) since the `Init` functions might
// depend on them.
func Init() {
	for _, cmd := range Commands {
		if err := cmd.Init(); err != nil {
			log.Fatal().Err(err).Msgf("failed to init command %v", cmd)
		}
	}
}
