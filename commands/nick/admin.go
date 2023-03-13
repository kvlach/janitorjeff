package nick

import (
	"github.com/janitorjeff/jeff-bot/commands/mask"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(m *core.Message) bool {
	return Advanced.Permitted(m)
}

func (admin) Names() []string {
	return Advanced.Names()
}

func (admin) Description() string {
	return Advanced.Description()
}

func (admin) UsageArgs() string {
	return Advanced.UsageArgs()
}

func (admin) Category() core.CommandCategory {
	return Advanced.Category()
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
	return AdvancedShow.Names()
}

func (adminShow) Description() string {
	return AdvancedShow.Description()
}

func (adminShow) UsageArgs() string {
	return AdvancedShow.UsageArgs()
}

func (c adminShow) Category() core.CommandCategory {
	return c.Parent().Category()
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
	nick, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return AdvancedShow.discord(nick, usrErr)
	default:
		return AdvancedShow.text(nick, usrErr)
	}
}

func (adminShow) core(m *core.Message) (string, error, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}

	t, usrErr, err := mask.Show(author)
	if usrErr != nil || err != nil {
		return "", usrErr, err
	}

	return Show(t.Person, t.Place)
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

func (adminSet) Names() []string {
	return AdvancedSet.Names()
}

func (adminSet) Description() string {
	return AdvancedSet.Description()
}

func (adminSet) UsageArgs() string {
	return AdvancedSet.UsageArgs()
}

func (c adminSet) Category() core.CommandCategory {
	return c.Parent().Category()
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
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	nick, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return AdvancedSet.discord(nick, usrErr)
	default:
		return AdvancedSet.text(nick, usrErr)
	}

}

func (adminSet) core(m *core.Message) (string, error, error) {
	nick := m.Command.Args[0]

	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}

	t, usrErr, err := mask.Show(author)
	if usrErr != nil || err != nil {
		return "", usrErr, err
	}

	usrErr, err = Set(nick, t.Person, t.Place)
	return nick, usrErr, err
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
	return AdvancedDelete.Names()
}

func (adminDelete) Description() string {
	return AdvancedDelete.Description()
}

func (adminDelete) UsageArgs() string {
	return AdvancedDelete.UsageArgs()
}

func (c adminDelete) Category() core.CommandCategory {
	return c.Parent().Category()
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
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return AdvancedDelete.discord(usrErr)
	default:
		return AdvancedDelete.text(usrErr)
	}
}

func (adminDelete) core(m *core.Message) (error, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return nil, err
	}

	t, usrErr, err := mask.Show(author)
	if usrErr != nil || err != nil {
		return usrErr, err
	}

	return Delete(t.Person, t.Place)
}
