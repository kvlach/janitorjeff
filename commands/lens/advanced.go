package lens

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"
	"strings"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.Message) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"lens",
	}
}

func (advanced) Description() string {
	return "A film calendar."
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
		AdvancedDirectors,
	}
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////////
//          //
// director //
//          //
//////////////

var AdvancedDirectors = advancedDirectors{}

type advancedDirectors struct{}

func (c advancedDirectors) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDirectors) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDirectors) Names() []string {
	return []string{
		"directors",
		"director",
	}
}

func (advancedDirectors) Description() string {
	return "Manage the list of directors being monitored."
}

func (c advancedDirectors) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedDirectors) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedDirectors) Examples() []string {
	return nil
}

func (advancedDirectors) Parent() core.CommandStatic {
	return Advanced
}

func (advancedDirectors) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedDirectorsAdd,
		AdvancedDirectorsDelete,
	}
}

func (advancedDirectors) Init() error {
	return nil
}

func (advancedDirectors) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

///////////////////
//               //
// directors add //
//               //
///////////////////

var AdvancedDirectorsAdd = advancedDirectorsAdd{}

type advancedDirectorsAdd struct{}

func (c advancedDirectorsAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDirectorsAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDirectorsAdd) Names() []string {
	return core.AliasesAdd
}

func (advancedDirectorsAdd) Description() string {
	return "Add a director to the list of directors being monitored for new releases."
}

func (advancedDirectorsAdd) UsageArgs() string {
	return "<name>"
}

func (c advancedDirectorsAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedDirectorsAdd) Examples() []string {
	return nil
}

func (advancedDirectorsAdd) Parent() core.CommandStatic {
	return AdvancedDirectors
}

func (advancedDirectorsAdd) Children() core.CommandsStatic {
	return nil
}

func (advancedDirectorsAdd) Init() error {
	return nil
}

func (c advancedDirectorsAdd) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedDirectorsAdd) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	name, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(name),
	}
	return embed, nil, nil
}

func (c advancedDirectorsAdd) text(m *core.Message) (string, core.Urr, error) {
	name, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(name), nil, nil
}

func (advancedDirectorsAdd) fmt(name string) string {
	return "Now also monitoring the director " + name
}

func (advancedDirectorsAdd) core(m *core.Message) (string, error) {
	// Use strings.Join instead of m.RawArgs to ensure that only one space
	// exists between each word.
	name := strings.Join(m.Command.Args, " ")
	return name, DirectorAdd(name)
}

//////////////////////
//                  //
// directors delete //
//                  //
//////////////////////

var AdvancedDirectorsDelete = advancedDirectorsDelete{}

type advancedDirectorsDelete struct{}

func (c advancedDirectorsDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDirectorsDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDirectorsDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedDirectorsDelete) Description() string {
	return "Delete a director from the list of directors being monitored for new releases."
}

func (advancedDirectorsDelete) UsageArgs() string {
	return "<name>"
}

func (c advancedDirectorsDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedDirectorsDelete) Examples() []string {
	return nil
}

func (advancedDirectorsDelete) Parent() core.CommandStatic {
	return AdvancedDirectors
}

func (advancedDirectorsDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedDirectorsDelete) Init() error {
	return nil
}

func (c advancedDirectorsDelete) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedDirectorsDelete) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	name, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(name),
	}
	return embed, nil, nil
}

func (c advancedDirectorsDelete) text(m *core.Message) (string, core.Urr, error) {
	name, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(name), nil, nil
}

func (advancedDirectorsDelete) fmt(name string) string {
	return "No longer monitoring the director " + name
}

func (advancedDirectorsDelete) core(m *core.Message) (string, error) {
	// Use strings.Join instead of m.RawArgs to ensure that only one space
	// exists between each word.
	name := strings.Join(m.Command.Args, " ")
	return name, DirectorDelete(name)
}
