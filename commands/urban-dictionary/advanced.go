package urban_dictionary

import (
	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.EventMessage) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"ud",
	}
}

func (advanced) Description() string {
	return "Search a term or get a random one on urban dictionary."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (advanced) Examples() []string {
	return nil
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedSearch,
		AdvancedRandom,
	}
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

////////////
//        //
// search //
//        //
////////////

var AdvancedSearch = advancedSearch{}

type advancedSearch struct{}

func (c advancedSearch) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSearch) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedSearch) Names() []string {
	return core.AliasesSearch
}

func (advancedSearch) Description() string {
	return "Search a term."
}

func (advancedSearch) UsageArgs() string {
	return "<term...>"
}

func (c advancedSearch) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSearch) Examples() []string {
	return nil
}

func (advancedSearch) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSearch) Children() core.CommandsStatic {
	return nil
}

func (advancedSearch) Init() error {
	return nil
}

func (c advancedSearch) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSearch) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	def, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return renderDiscord(def, urr), urr, nil
}

func (c advancedSearch) text(m *core.EventMessage) (string, core.Urr, error) {
	def, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return renderText(def, urr), urr, nil
}

func (advancedSearch) core(m *core.EventMessage) (definition, core.Urr, error) {
	term := m.RawArgs(0)
	return Search(term)
}

////////////
//        //
// random //
//        //
////////////

var AdvancedRandom = advancedRandom{}

type advancedRandom struct{}

func (c advancedRandom) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRandom) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedRandom) Names() []string {
	return []string{
		"random",
		"rand",
	}
}

func (advancedRandom) Description() string {
	return "Get a random term."
}

func (advancedRandom) UsageArgs() string {
	return ""
}

func (c advancedRandom) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRandom) Examples() []string {
	return nil
}

func (advancedRandom) Parent() core.CommandStatic {
	return Advanced
}

func (advancedRandom) Children() core.CommandsStatic {
	return nil
}

func (advancedRandom) Init() error {
	return nil
}

func (c advancedRandom) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord()
	default:
		return c.text()
	}
}

func (c advancedRandom) discord() (*dg.MessageEmbed, core.Urr, error) {
	def, urr, err := c.core()
	if err != nil {
		return nil, nil, err
	}
	return renderDiscord(def, urr), urr, nil
}

func (c advancedRandom) text() (string, core.Urr, error) {
	def, urr, err := c.core()
	if err != nil {
		return "", nil, err
	}
	return renderText(def, urr), urr, nil
}

func (advancedRandom) core() (definition, core.Urr, error) {
	return Random()
}
