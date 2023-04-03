package mask

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.Message) bool {
	return true
}

func (admin) Names() []string {
	return []string{
		"mask",
	}
}

func (admin) Description() string {
	return "Execute commands as if you are a person in a place."
}

func (c admin) UsageArgs() string {
	return c.Children().Usage()
}

func (admin) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (admin) Examples() []string {
	return nil
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminShow,
		AdminSet,
		AdminDelete,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

//////////
//      //
// show //
//      //
//////////

var AdminShow = adminShow{}

type adminShow struct{}

func (c adminShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminShow) Names() []string {
	return core.AliasesShow
}

func (adminShow) Description() string {
	return "Show your current mask."
}

func (adminShow) UsageArgs() string {
	return ""
}

func (c adminShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminShow) Examples() []string {
	return nil
}

func (adminShow) Parent() core.CommandStatic {
	return Admin
}

func (adminShow) Children() core.CommandsStatic {
	return nil
}

func (adminShow) Init() error {
	return nil
}

func (c adminShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, nil
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, t),
	}
	return embed, usrErr, nil
}

func (c adminShow) text(m *core.Message) (string, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, nil
	}
	return c.err(usrErr, t), usrErr, nil
}

func (adminShow) err(usrErr error, t Target) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("person=%d place=%d", t.Person, t.Place)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminShow) core(m *core.Message) (Target, error, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return Target{}, nil, err
	}
	return Show(author)
}

/////////
//     //
// set //
//     //
/////////

var AdminSet = adminSet{}

type adminSet struct{}

func (c adminSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (c adminSet) Names() []string {
	return []string{
		"set",
	}
}

func (adminSet) Description() string {
	return "Set your mask."
}

func (adminSet) UsageArgs() string {
	return "<person> <place>"
}

func (c adminSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminSet) Examples() []string {
	return nil
}

func (adminSet) Parent() core.CommandStatic {
	return Admin
}

func (adminSet) Children() core.CommandsStatic {
	return nil
}

func (adminSet) Init() error {
	return nil
}

func (c adminSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminSet) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: fmt.Sprintf("person=%d place=%d", t.Person, t.Place),
	}
	return embed, nil, nil
}

func (c adminSet) text(m *core.Message) (string, error, error) {
	t, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("person=%d place=%d", t.Person, t.Place), nil, nil
}

func (adminSet) core(m *core.Message) (Target, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return Target{}, err
	}
	userID := m.Command.Args[0]
	locID := m.Command.Args[1]
	return Set(m.Client, author, userID, locID)
}

////////////
//        //
// delete //
//        //
////////////

var AdminDelete = adminDelete{}

type adminDelete struct{}

func (c adminDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminDelete) Names() []string {
	return core.AliasesDelete
}

func (adminDelete) Description() string {
	return "Delete your current mask."
}

func (adminDelete) UsageArgs() string {
	return ""
}

func (c adminDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminDelete) Examples() []string {
	return nil
}

func (adminDelete) Parent() core.CommandStatic {
	return Admin
}

func (adminDelete) Children() core.CommandsStatic {
	return nil
}

func (adminDelete) Init() error {
	return nil
}

func (c adminDelete) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c adminDelete) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (adminDelete) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Deleted your mask."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminDelete) core(m *core.Message) (error, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return nil, err
	}
	return Delete(author), nil
}
