package commands

import (
	"github.com/janitorjeff/jeff-bot/commands/category"
	"github.com/janitorjeff/jeff-bot/commands/command"
	"github.com/janitorjeff/jeff-bot/commands/connect"
	"github.com/janitorjeff/jeff-bot/commands/help"
	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/commands/paintball"
	"github.com/janitorjeff/jeff-bot/commands/prefix"
	"github.com/janitorjeff/jeff-bot/commands/rps"
	"github.com/janitorjeff/jeff-bot/commands/scope"
	"github.com/janitorjeff/jeff-bot/commands/test"
	"github.com/janitorjeff/jeff-bot/commands/time"
	"github.com/janitorjeff/jeff-bot/commands/title"
	"github.com/janitorjeff/jeff-bot/commands/urban-dictionary"
	"github.com/janitorjeff/jeff-bot/commands/wikipedia"
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

var Commands = core.CommandsStatic{
	category.Normal,
	category.Advanced,

	command.Advanced,

	connect.Normal,

	help.Normal,
	help.Advanced,
	help.Admin,

	nick.Normal,
	nick.Advanced,
	nick.Admin,

	paintball.Normal,

	prefix.Normal,
	prefix.Advanced,
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
