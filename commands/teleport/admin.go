package teleport

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(m *core.Message) bool {
	return true
}

func (admin) Names() []string {
	return []string{
		"teleport",
		"tp",
	}
}

func (admin) Description() string {
	return "Teleport to a place and back. Execute commands as if you were there."
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
		AdminTo,
		AdminHome,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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
	return "Show the current teleport status."
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

func (c adminShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Type:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	place, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(place),
	}
	return embed, nil, nil
}

func (c adminShow) text(m *core.Message) (string, core.Urr, error) {
	place, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(place), nil, nil
}

func (adminShow) fmt(place core.Place) string {
	zero := core.Place{}
	if place == zero {
		return "Home sweet home."
	}
	return fmt.Sprintf("Currently teleported to: exact=%d logical=%d.", place.Exact, place.Logical)
}

func (adminShow) core(m *core.Message) (core.Place, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return core.Place{}, err
	}
	place, ok := core.Teleports.Get(author)
	if !ok {
		return core.Place{}, nil
	}
	return place, nil
}

////////
//    //
// to //
//    //
////////

var AdminTo = adminTo{}

type adminTo struct{}

func (c adminTo) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminTo) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminTo) Names() []string {
	return []string{
		"to",
		"->",
	}
}

func (adminTo) Description() string {
	return "Teleport to a place."
}

func (adminTo) UsageArgs() string {
	return "<frontend> <location>"
}

func (c adminTo) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminTo) Examples() []string {
	return nil
}

func (adminTo) Parent() core.CommandStatic {
	return Admin
}

func (adminTo) Children() core.CommandsStatic {
	return nil
}

func (adminTo) Init() error {
	return nil
}

func (c adminTo) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Type:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminTo) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	to, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt("**"+to+"**", urr),
	}
	return embed, urr, nil
}

func (c adminTo) text(m *core.Message) (string, core.Urr, error) {
	to, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(to, urr), urr, nil
}

func (adminTo) fmt(to string, urr core.Urr) string {
	switch urr {
	case nil:
		return "Teleported to " + to + "."
	default:
		return urr.Error()
	}
}

func (adminTo) core(m *core.Message) (string, core.Urr, error) {
	id := m.Command.Args[1]

	f, urr := core.Frontends.Match(m.Command.Args[0])
	if urr != nil {
		return "", urr, nil
	}

	exact, err := f.PlaceExact(id)
	if err != nil {
		return "", nil, err
	}
	logical, err := f.PlaceLogical(id)
	if err != nil {
		return "", nil, err
	}

	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}
	core.Teleports.Set(author, core.Place{
		Exact:   exact,
		Logical: logical,
	})
	return id, nil, nil
}

//////////
//      //
// home //
//      //
//////////

var AdminHome = adminHome{}

type adminHome struct{}

func (c adminHome) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminHome) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminHome) Names() []string {
	return []string{
		"home",
		"back",
		"<-",
	}
}

func (adminHome) Description() string {
	return "Teleport back from a place."
}

func (adminHome) UsageArgs() string {
	return ""
}

func (c adminHome) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminHome) Examples() []string {
	return nil
}

func (adminHome) Parent() core.CommandStatic {
	return Admin
}

func (adminHome) Children() core.CommandsStatic {
	return nil
}

func (adminHome) Init() error {
	return nil
}

func (c adminHome) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Type:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminHome) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c adminHome) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (adminHome) fmt() string {
	return "Teleported back home."
}

func (adminHome) core(m *core.Message) error {
	author, err := m.Author.Scope()
	if err != nil {
		return err
	}
	core.Teleports.Delete(author)
	return nil
}
