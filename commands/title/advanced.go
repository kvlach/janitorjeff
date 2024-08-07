package title

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/twitch"

	"github.com/rs/zerolog/log"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.EventMessage) bool {
	if m.Frontend.Type() != twitch.Frontend.Type() {
		return false
	}
	mod, err := m.Author.Moderator()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if author is mod")
		return false
	}
	return mod
}

func (advanced) Names() []string {
	return []string{
		"title",
	}
}

func (advanced) Description() string {
	return "Show or edit the current title."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryModerators
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
		AdvancedEdit,
	}
}

func (advanced) Init() error {
	return nil
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
	return "Show the current title."
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
	switch m.Frontend.Type() {
	case twitch.Frontend.Type():
		return c.twitch(m)
	default:
		panic("this should never happen")
	}
}

func (advancedShow) twitch(m *core.EventMessage) (string, core.Urr, error) {
	hx, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}
	hix, err := m.Here.IDExact()
	if err != nil {
		return "", nil, err
	}
	t, err := hx.GetTitle(hix)
	return t, nil, err
}

//////////
//      //
// edit //
//      //
//////////

var AdvancedEdit = advancedEdit{}

type advancedEdit struct{}

func (c advancedEdit) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedEdit) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedEdit) Names() []string {
	return core.AliasesEdit
}

func (advancedEdit) Description() string {
	return "Edit the current title."
}

func (advancedEdit) UsageArgs() string {
	return "<title...>"
}

func (c advancedEdit) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedEdit) Examples() []string {
	return nil
}

func (advancedEdit) Parent() core.CommandStatic {
	return Advanced
}

func (advancedEdit) Children() core.CommandsStatic {
	return nil
}

func (advancedEdit) Init() error {
	return nil
}

func (c advancedEdit) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case twitch.Frontend.Type():
		return c.twitch(m)
	default:
		panic("this should never happen")
	}
}

func (advancedEdit) twitch(m *core.EventMessage) (string, core.Urr, error) {
	hx, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}

	title := m.RawArgs(0)

	hix, err := m.Here.IDExact()
	if err != nil {
		return "", nil, err
	}

	urr, err := hx.SetTitle(hix, title)
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("Title set to: %s", title), nil, nil
}
