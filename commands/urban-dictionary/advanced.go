package urban_dictionary

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = &core.CommandStatic{
	Names: []string{
		"ud",
	},
	Description: "Search a term or get a random one on urban dictionary.",
	UsageArgs:   "(search | random)",
	Frontends:   frontends.All,
	Run:         advancedRun,

	Children: core.Commands{
		{
			Names: []string{
				"search",
				"find",
			},
			Description: "Search a term.",
			UsageArgs:   "<term...>",
			Run:         advancedRunSearch,
		},
		{
			Names: []string{
				"random",
				"rand",
			},
			Description: "Get a random term.",
			UsageArgs:   "",
			Run:         advancedRunRandom,
		},
	},
}

func advancedRun(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////
//        //
// search //
//        //
////////////

func advancedRunSearch(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return advancedRunSearchDiscord(m)
	default:
		return advancedRunSearchText(m)
	}
}

func advancedRunSearchDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	def, usrErr, err := advancedRunSearchCore(m)
	if err != nil {
		return nil, nil, err
	}
	return renderDiscord(def, usrErr), usrErr, nil
}

func advancedRunSearchText(m *core.Message) (string, error, error) {
	def, usrErr, err := advancedRunSearchCore(m)
	if err != nil {
		return "", nil, err
	}
	return renderText(def, usrErr), usrErr, nil
}

func advancedRunSearchCore(m *core.Message) (definition, error, error) {
	term := m.RawArgs(0)
	return search(term)
}

////////////
//        //
// random //
//        //
////////////

func advancedRunRandom(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return advancedRunRandomDiscord()
	default:
		return advancedRunRandomText()
	}
}

func advancedRunRandomDiscord() (*dg.MessageEmbed, error, error) {
	def, usrErr, err := advancedRunRandomCore()
	if err != nil {
		return nil, nil, err
	}
	return renderDiscord(def, usrErr), usrErr, nil
}

func advancedRunRandomText() (string, error, error) {
	def, usrErr, err := advancedRunRandomCore()
	if err != nil {
		return "", nil, err
	}
	return renderText(def, usrErr), usrErr, nil
}

func advancedRunRandomCore() (definition, error, error) {
	return rand()
}
