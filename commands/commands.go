package commands

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/commands/audio"
	"github.com/janitorjeff/jeff-bot/commands/category"
	"github.com/janitorjeff/jeff-bot/commands/connect"
	"github.com/janitorjeff/jeff-bot/commands/custom-command"
	"github.com/janitorjeff/jeff-bot/commands/help"
	"github.com/janitorjeff/jeff-bot/commands/id"
	"github.com/janitorjeff/jeff-bot/commands/mask"
	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/commands/paintball"
	"github.com/janitorjeff/jeff-bot/commands/prefix"
	"github.com/janitorjeff/jeff-bot/commands/rps"
	"github.com/janitorjeff/jeff-bot/commands/time"
	"github.com/janitorjeff/jeff-bot/commands/title"
	"github.com/janitorjeff/jeff-bot/commands/urban-dictionary"
	"github.com/janitorjeff/jeff-bot/commands/wikipedia"
	"github.com/janitorjeff/jeff-bot/commands/youtube"
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

var Commands = core.CommandsStatic{
	audio.Advanced,

	category.Normal,
	category.Advanced,

	connect.Normal,

	custom_command.Advanced,

	help.Normal,
	help.Advanced,
	help.Admin,

	id.Normal,

	mask.Admin,

	nick.Normal,
	nick.Advanced,
	nick.Admin,

	paintball.Normal,

	prefix.Normal,
	prefix.Advanced,
	prefix.Admin,

	rps.Normal,

	time.Normal,
	time.Advanced,

	title.Normal,
	title.Advanced,

	urban_dictionary.Normal,
	urban_dictionary.Advanced,

	wikipedia.Normal,

	youtube.Normal,
	youtube.Advanced,
}

// checkNameCollisions checks if there are any name is used more than once in
// the given list of commands
func checkNameCollisions(cmds core.CommandsStatic) {
	// will act like a set
	names := map[string]struct{}{}

	for _, cmd := range cmds {
		for _, n := range cmd.Names() {
			if _, ok := names[n]; ok {
				panic(fmt.Sprintf("name %s already exists in %v", n, names))
			}
			names[n] = struct{}{}
		}
	}
}

// checkDifferentParentType will check if all the parent's children are of the
// same command type as the parent.
func checkDifferentParentType(parent core.CommandStatic, children core.CommandsStatic) {
	for _, child := range children {
		if parent.Type() != child.Type() {
			panic(fmt.Sprintf("child %v is of different type from parent %v", child.Names(), parent.Names()))
		}
	}
}

// recurse will check if the children have set their parents correctly
// recurse will recursively go through all of the given commands children and
// will check if there's any name collisions and if the children have set
// their parents correctly
func recurse(cmd core.CommandStatic) {
	if cmd.Children() == nil {
		return
	}

	checkNameCollisions(cmd.Children())
	checkDifferentParentType(cmd, cmd.Children())

	for _, child := range cmd.Children() {
		if child.Parent() != cmd {
			panic(fmt.Sprintf("incorrect parent-child relationship, expected parent %v for child %v but got %v", cmd.Names(), child.Names(), child.Parent().Names()))
		}
		recurse(child)
	}
}

func init() {
	for _, cmd := range Commands {
		recurse(cmd)
	}
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
