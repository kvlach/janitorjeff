package nick

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

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
		"nick",
		"nickname",
	}
}

func (advanced) Description() string {
	return "Set, view or delete your nickname."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
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
	return core.DB.Init(dbSchema)
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

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
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

func (c advancedShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedShow) Names() []string {
	return core.Show
}

func (advancedShow) Description() string {
	return "View your current nickname."
}

func (advancedShow) UsageArgs() string {
	return ""
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

func (c advancedShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, nick),
	}

	return embed, usrErr, nil
}

func (c advancedShow) text(m *core.Message) (string, error, error) {
	nick, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return c.err(usrErr, nick), usrErr, nil
}

func (advancedShow) err(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your nickname is: %s", nick)
	case errPersonNotFound:
		return "You have not set a nickname."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedShow) core(m *core.Message) (string, error, error) {
	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
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

func (c advancedSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSet) Names() []string {
	return []string{
		"set",
	}
}

func (advancedSet) Description() string {
	return "Set your nickname."
}

func (advancedSet) UsageArgs() string {
	return "<nickname>"
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

func (c advancedSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSet) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, nick),
	}

	return embed, usrErr, nil
}

func (c advancedSet) text(m *core.Message) (string, error, error) {
	nick, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return c.err(usrErr, nick), usrErr, nil
}

func (c advancedSet) err(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Nickname set to %s", nick)
	case errNickExists:
		return fmt.Sprintf("Nickname %s is already being used by another user.", nick)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (c advancedSet) core(m *core.Message) (string, error, error) {
	nick := m.Command.Args[0]

	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	usrErr, err := Set(nick, author, here)
	return nick, usrErr, err
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

func (c advancedDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDelete) Names() []string {
	return core.Delete
}

func (advancedDelete) Description() string {
	return "Delete your nickname."
}

func (advancedDelete) UsageArgs() string {
	return ""
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

func (c advancedDelete) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}

	return embed, usrErr, nil
}

func (c advancedDelete) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedDelete) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Deleted your nickname."
	case errPersonNotFound:
		return "Can't delete, you haven't set a nickname here."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedDelete) core(m *core.Message) (error, error) {
	author, err := m.Author()
	if err != nil {
		return nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	return Delete(author, here)
}
