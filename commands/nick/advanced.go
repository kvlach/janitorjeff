package nick

import (
	"fmt"

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
		"nick",
		"nickname",
	}
}

func (advanced) Description() string {
	return "Show, set or delete your nickname."
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
		AdvancedShow,
		AdvancedSet,
		AdvancedDelete,
	}
}

func (c advanced) Init() error {
	c.discordAppCommand()
	return nil
}

func (advanced) discordAppCommand() {
	cmd := &dg.ApplicationCommand{
		Name:        Advanced.Names()[0],
		Type:        dg.ChatApplicationCommand,
		Description: Advanced.Description(),
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        AdvancedShow.Names()[0],
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: AdvancedShow.Description(),
			},
			{
				Name:        AdvancedSet.Names()[0],
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: AdvancedSet.Description(),
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "nickname",
						Type:        dg.ApplicationCommandOptionString,
						Description: "give nickname",
						Required:    true,
					},
				},
			},
			{
				Name:        AdvancedDelete.Names()[0],
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: AdvancedDelete.Description(),
			},
		},
	}

	discord.RegisterAppCommand(cmd)
}

func (advanced) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////
//      //
// show //
//      //
//////////

var AdvancedShow = advancedShow{}

type advancedShow struct{}

func (c advancedShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedShow) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedShow) Names() []string {
	return core.AliasesShow
}

func (advancedShow) Description() string {
	return "Show your current nickname."
}

func (advancedShow) UsageArgs() string {
	return ""
}

func (c advancedShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedShow) Examples() []string {
	return nil
}

func (advancedShow) Parent() core.CommandStatic {
	return Advanced
}

func (advancedShow) Children() core.CommandsStatic {
	return nil
}

func (advancedShow) Init() error {
	return nil
}

func (c advancedShow) Run(m *core.EventMessage) (any, core.Urr, error) {
	nick, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(nick, urr)
	default:
		return c.text(nick, urr)
	}
}

func (c advancedShow) discord(nick string, urr core.Urr) (*dg.MessageEmbed, core.Urr, error) {
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, fmt.Sprintf("**%s**", nick)),
	}
	return embed, urr, nil
}

func (c advancedShow) text(nick string, urr core.Urr) (string, core.Urr, error) {
	nick = fmt.Sprintf("'%s'", nick)
	return c.fmt(urr, nick), urr, nil
}

func (advancedShow) fmt(urr core.Urr, nick string) string {
	switch urr {
	case nil:
		return fmt.Sprintf("Your nickname is: %s", nick)
	case core.UrrValNil:
		return "You have not set a nickname."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedShow) core(m *core.EventMessage) (string, core.Urr, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", nil, err
	}

	return Show(author, here)
}

/////////
//     //
// set //
//     //
/////////

var AdvancedSet = advancedSet{}

type advancedSet struct{}

func (c advancedSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSet) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedSet) Names() []string {
	return core.AliasesSet
}

func (advancedSet) Description() string {
	return "Set your nickname."
}

func (advancedSet) UsageArgs() string {
	return "<nickname>"
}

func (c advancedSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSet) Examples() []string {
	return nil
}

func (advancedSet) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSet) Children() core.CommandsStatic {
	return nil
}

func (advancedSet) Init() error {
	return nil
}

func (c advancedSet) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	nick, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(nick, urr)
	default:
		return c.text(nick, urr)
	}
}

func (c advancedSet) discord(nick string, urr error) (*dg.MessageEmbed, core.Urr, error) {
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, fmt.Sprintf("**%s**", nick)),
	}
	return embed, urr, nil
}

func (c advancedSet) text(nick string, urr error) (string, core.Urr, error) {
	nick = fmt.Sprintf("'%s'", nick)
	return c.fmt(urr, nick), urr, nil
}

func (c advancedSet) fmt(urr core.Urr, nick string) string {
	switch urr {
	case nil:
		return fmt.Sprintf("Nickname set to %s", nick)
	default:
		return fmt.Sprint(urr)
	}
}

func (c advancedSet) core(m *core.EventMessage) (string, core.Urr, error) {
	nick := m.Command.Args[0]

	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", nil, err
	}

	urr, err := Set(nick, author, here)
	return nick, urr, err
}

////////////
//        //
// delete //
//        //
////////////

var AdvancedDelete = advancedDelete{}

type advancedDelete struct{}

func (c advancedDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDelete) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedDelete) Description() string {
	return "Delete your nickname."
}

func (advancedDelete) UsageArgs() string {
	return ""
}

func (c advancedDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedDelete) Examples() []string {
	return nil
}

func (advancedDelete) Parent() core.CommandStatic {
	return Advanced
}

func (advancedDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedDelete) Init() error {
	return nil
}

func (c advancedDelete) Run(m *core.EventMessage) (any, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord()
	default:
		return c.text()
	}
}

func (c advancedDelete) discord() (*dg.MessageEmbed, core.Urr, error) {
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedDelete) text() (string, core.Urr, error) {
	return c.fmt(), nil, nil
}

func (advancedDelete) fmt() string {
	return "Deleted your nickname."
}

func (advancedDelete) core(m *core.EventMessage) core.Urr {
	author, err := m.Author.Scope()
	if err != nil {
		return err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return Delete(author, here)
}
