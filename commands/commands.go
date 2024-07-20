package commands

import (
	"fmt"
	"net/http"

	"github.com/kvlach/janitorjeff/commands/audio"
	"github.com/kvlach/janitorjeff/commands/category"
	"github.com/kvlach/janitorjeff/commands/connect"
	"github.com/kvlach/janitorjeff/commands/custom-command"
	"github.com/kvlach/janitorjeff/commands/discord"
	"github.com/kvlach/janitorjeff/commands/god"
	"github.com/kvlach/janitorjeff/commands/help"
	"github.com/kvlach/janitorjeff/commands/id"
	"github.com/kvlach/janitorjeff/commands/lens"
	"github.com/kvlach/janitorjeff/commands/mask"
	"github.com/kvlach/janitorjeff/commands/nick"
	"github.com/kvlach/janitorjeff/commands/paintball"
	"github.com/kvlach/janitorjeff/commands/prefix"
	"github.com/kvlach/janitorjeff/commands/rps"
	"github.com/kvlach/janitorjeff/commands/search"
	"github.com/kvlach/janitorjeff/commands/streak"
	"github.com/kvlach/janitorjeff/commands/teleport"
	"github.com/kvlach/janitorjeff/commands/time"
	"github.com/kvlach/janitorjeff/commands/title"
	"github.com/kvlach/janitorjeff/commands/twitch"
	"github.com/kvlach/janitorjeff/commands/urban-dictionary"
	"github.com/kvlach/janitorjeff/commands/wikipedia"
	"github.com/kvlach/janitorjeff/commands/youtube"
	"github.com/kvlach/janitorjeff/core"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var Commands = core.CommandsStatic{
	audio.Advanced,

	category.Normal,
	category.Advanced,

	connect.Normal,

	custom_command.Advanced,

	discord.Admin,

	god.Advanced,
	god.Normal,
	god.Admin,

	help.Normal,
	help.Advanced,
	help.Admin,

	id.Normal,

	lens.Advanced,

	mask.Admin,

	nick.Normal,
	nick.Advanced,
	nick.Admin,

	paintball.Normal,

	prefix.Normal,
	prefix.Advanced,
	prefix.Admin,

	rps.Normal,

	search.Advanced,

	streak.Admin,
	streak.Advanced,
	streak.Normal,

	teleport.Admin,

	time.Advanced,
	time.NormalTime,
	time.NormalTimezone,

	title.Normal,
	title.Advanced,

	twitch.Admin,

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

// checkDifferentParentType will check if all the parent's children are of the
// same command type as the parent.
func checkDifferentParentCategory(parent core.CommandStatic, children core.CommandsStatic) {
	for _, child := range children {
		if parent.Category() != child.Category() {
			panic(fmt.Sprintf("child %v is of different category from parent %v", child.Names(), parent.Names()))
		}
	}
}

// checkWrongParentChild will check if a parent's children have set their parent
// correctly.
func checkWrongParentInChild(parent core.CommandStatic, children core.CommandsStatic) {
	for _, child := range children {
		if child.Parent() != parent {
			panic(fmt.Sprintf("incorrect parent-child relationship, expected parent %v for child %v but got %v", parent.Names(), child.Names(), child.Parent().Names()))
		}
	}
}

// recurse will recursively go through all the children of the passed command
// and perform the following checks on them:
//
//   - Name collisions: will check if a name is used more than once among the
//     children
//   - Wrong type: Will check if a child's command type is different from its
//     parent's
//   - Wrong parent: Will check if a child has set its parent incorrectly
//
// If any of the above checks fail the program will immediately panic.
func recurse(cmd core.CommandStatic) {
	if cmd.Children() == nil {
		return
	}

	checkNameCollisions(cmd.Children())
	checkDifferentParentType(cmd, cmd.Children())
	checkDifferentParentCategory(cmd, cmd.Children())
	checkWrongParentInChild(cmd, cmd.Children())

	for _, child := range cmd.Children() {
		recurse(child)
	}
}

func init() {
	for _, cmd := range Commands {
		recurse(cmd)
	}
}

// Init must be run after all the global variables have been set (including ones
// that frontend init functions might set) since the `Init` functions might
// depend on them.
func Init() {
	for _, cmd := range Commands {
		if err := cmd.Init(); err != nil {
			log.Fatal().Err(err).Msgf("failed to init command %v", core.Format(cmd, "!"))
		}
	}
}

type CommandJSON struct {
	Names       []string `json:"names"`
	Description string   `json:"description"`
	Example     string   `json:"example"`
	Parent      int      `json:"parent"`
	Children    []int    `json:"children"`
	Category    string   `json:"category"`
}

type Resp struct {
	Prefix     string        `json:"prefix"`
	Categories []string      `json:"categories"`
	Commands   []CommandJSON `json:"commands"`
}

func ToJSON(t core.CommandType) Resp {
	resp := Resp{
		Prefix:     "!",
		Categories: []string{},
		Commands:   []CommandJSON{},
	}

	commandsIndex := map[core.CommandStatic]int{}

	categories := map[core.CommandCategory]struct{}{}

	Commands.Recurse(func(cmd core.CommandStatic) {
		if cmd.Type() != t {
			return
		}

		cmdJSON := CommandJSON{
			Names:       cmd.Names(),
			Description: cmd.Description(),
			Example:     "Example.",
			Category:    string(cmd.Category()),
		}

		if cmd.Parent() == nil {
			cmdJSON.Parent = -1
		} else {
			cmdJSON.Parent = commandsIndex[cmd.Parent()]
		}

		if cmd.Children() == nil {
			cmdJSON.Children = []int{}
		}

		commandsIndex[cmd] = len(resp.Commands)
		resp.Commands = append(resp.Commands, cmdJSON)

		categories[cmd.Category()] = struct{}{}
	})

	Commands.Recurse(func(cmd core.CommandStatic) {
		if cmd.Type() != t {
			return
		}

		if cmd.Children() == nil {
			return
		}

		index := commandsIndex[cmd]

		for _, child := range cmd.Children() {
			resp.Commands[index].Children = append(resp.Commands[index].Children, commandsIndex[child])
		}
	})

	for cat := range categories {
		resp.Categories = append(resp.Categories, string(cat))
	}

	return resp
}

func init() {
	core.Gin.GET("/api/v1/commands", func(c *gin.Context) {
		var t core.CommandType
		switch c.DefaultQuery("type", "normal") {
		case "normal":
			t = core.Normal
		case "advanced":
			t = core.Advanced
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
			return
		}
		c.JSON(http.StatusOK, ToJSON(t))
	})
}
