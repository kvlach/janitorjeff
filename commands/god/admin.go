package god

import (
	"fmt"
	"strconv"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.EventMessage) bool {
	return true
}

func (admin) Names() []string {
	return Advanced.Names()
}

func (admin) Description() string {
	return Advanced.Description()
}

func (c admin) UsageArgs() string {
	return c.Children().Usage()
}

func (admin) Category() core.CommandCategory {
	return Advanced.Category()
}

func (admin) Examples() []string {
	return nil
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminMax,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

/////////
//     //
// max //
//     //
/////////

var AdminMax = adminMax{}

type adminMax struct{}

func (c adminMax) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminMax) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminMax) Names() []string {
	return []string{
		"max",
	}
}

func (adminMax) Description() string {
	return "Determine the max number of tokens the responses can contain."
}

func (c adminMax) UsageArgs() string {
	return c.Children().Usage()
}

func (c adminMax) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminMax) Examples() []string {
	return nil
}

func (adminMax) Parent() core.CommandStatic {
	return Admin
}

func (adminMax) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminMaxShow,
		AdminMaxSet,
	}
}

func (adminMax) Init() error {
	return nil
}

func (adminMax) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////////
//          //
// max show //
//          //
//////////////

var AdminMaxShow = adminMaxShow{}

type adminMaxShow struct{}

func (c adminMaxShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminMaxShow) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminMaxShow) Names() []string {
	return core.AliasesShow
}

func (adminMaxShow) Description() string {
	return "Show the max length a response can have."
}

func (adminMaxShow) UsageArgs() string {
	return ""
}

func (c adminMaxShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminMaxShow) Examples() []string {
	return nil
}

func (adminMaxShow) Parent() core.CommandStatic {
	return AdminMax
}

func (adminMaxShow) Children() core.CommandsStatic {
	return nil
}

func (adminMaxShow) Init() error {
	return nil
}

func (c adminMaxShow) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminMaxShow) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	max, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(fmt.Sprintf("**%d**", max)),
	}
	return embed, nil, nil
}

func (c adminMaxShow) text(m *core.EventMessage) (string, core.Urr, error) {
	max, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(strconv.Itoa(max)), nil, nil
}

func (adminMaxShow) fmt(max string) string {
	return fmt.Sprintf("The response will contain a maximum of %s tokens.", max)
}

func (adminMaxShow) core(m *core.EventMessage) (int, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return 0, err
	}
	return MaxGet(here)
}

/////////////
//         //
// max set //
//         //
/////////////

var AdminMaxSet = adminMaxSet{}

type adminMaxSet struct{}

func (c adminMaxSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminMaxSet) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminMaxSet) Names() []string {
	return core.AliasesSet
}

func (adminMaxSet) Description() string {
	return "Set the max length a response can have."
}

func (adminMaxSet) UsageArgs() string {
	return "<length:int>"
}

func (c adminMaxSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminMaxSet) Examples() []string {
	return nil
}

func (adminMaxSet) Parent() core.CommandStatic {
	return AdminMax
}

func (adminMaxSet) Children() core.CommandsStatic {
	return nil
}

func (adminMaxSet) Init() error {
	return nil
}

func (c adminMaxSet) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 0 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c adminMaxSet) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	max, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(
			fmt.Sprintf("**%d**", max),
			"**"+m.Command.Args[0]+"**",
			urr,
		),
	}
	return embed, nil, nil
}

func (c adminMaxSet) text(m *core.EventMessage) (string, core.Urr, error) {
	max, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(strconv.Itoa(max), m.Command.Args[0], urr), nil, nil
}

func (adminMaxSet) fmt(max, arg string, urr core.Urr) string {
	switch urr {
	case nil:
		return "Updated max tokens to " + max + "."
	case UrrNotInt:
		return "Expected integer, got " + arg + "."
	default:
		return urr.Error()
	}
}

func (adminMaxSet) core(m *core.EventMessage) (int, core.Urr, error) {
	max, err := strconv.Atoi(m.Command.Args[0])
	if err != nil {
		return 0, UrrNotInt, nil
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return 0, nil, err
	}
	return max, nil, MaxSet(here, max)
}
