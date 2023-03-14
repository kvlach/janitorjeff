package commands

import (
	"fmt"
	"net/http"

	"github.com/janitorjeff/jeff-bot/commands/audio"
	"github.com/janitorjeff/jeff-bot/commands/category"
	"github.com/janitorjeff/jeff-bot/commands/connect"
	"github.com/janitorjeff/jeff-bot/commands/custom-command"
	"github.com/janitorjeff/jeff-bot/commands/god"
	"github.com/janitorjeff/jeff-bot/commands/help"
	"github.com/janitorjeff/jeff-bot/commands/id"
	"github.com/janitorjeff/jeff-bot/commands/mask"
	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/commands/paintball"
	"github.com/janitorjeff/jeff-bot/commands/prefix"
	"github.com/janitorjeff/jeff-bot/commands/rps"
	"github.com/janitorjeff/jeff-bot/commands/search"
	"github.com/janitorjeff/jeff-bot/commands/time"
	"github.com/janitorjeff/jeff-bot/commands/title"
	"github.com/janitorjeff/jeff-bot/commands/tts"
	"github.com/janitorjeff/jeff-bot/commands/urban-dictionary"
	"github.com/janitorjeff/jeff-bot/commands/wikipedia"
	"github.com/janitorjeff/jeff-bot/commands/youtube"
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var Commands = core.CommandsStatic{
	audio.Advanced,

	category.Normal,
	category.Advanced,

	connect.Normal,

	custom_command.Advanced,

	god.Advanced,
	god.Normal,

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

	search.Advanced,

	time.Normal,
	time.Advanced,

	title.Normal,
	title.Advanced,

	tts.Advanced,
	tts.NormalTTS,
	tts.NormalVoice,
	tts.NormalSubOnly,

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

type CommandJSON struct {
	Names       []string `json:"names"`
	Description string   `json:"description"`
	Example     string   `json:"example"`
}

type CategoryJSON struct {
	Name     string        `json:"name"`
	Commands []CommandJSON `json:"commands"`
}

type Resp struct {
	Prefix     string         `json:"prefix"`
	Categories []CategoryJSON `json:"categories"`
}

func init() {
	r := gin.Default()

	r.GET("/api/v1/commands", func(c *gin.Context) {
		resp := Resp{
			Prefix:     "!",
			Categories: []CategoryJSON{},
		}

		for _, cmd := range Commands {
			if cmd.Type() != core.Normal {
				continue
			}

			cmdJson := CommandJSON{
				Names:       cmd.Names(),
				Description: cmd.Description(),
				Example:     "Example.",
			}

			exists := false
			for _, cat := range resp.Categories {
				if cat.Name == string(cmd.Category()) {
					exists = true
					break
				}
			}

			if !exists {
				resp.Categories = append(resp.Categories, CategoryJSON{
					Name:     string(cmd.Category()),
					Commands: []CommandJSON{},
				})
			}

			index := 0
			for i, cat := range resp.Categories {
				if cat.Name == string(cmd.Category()) {
					index = i
				}
			}

			resp.Categories[index].Commands = append(resp.Categories[index].Commands, cmdJson)
		}

		c.IndentedJSON(http.StatusOK, resp)
		//c.JSON(http.StatusOK, resp)
	})

	//r.Run("localhost:" + "13420")
}
