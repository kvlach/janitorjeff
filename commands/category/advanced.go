package category

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
		"category",
		"game",
	}
}

func (advanced) Description() string {
	return "Show or edit the current category."
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
	return "Show the current category."
}

func (advancedShow) UsageArgs() string {
	return ""
}

func (c advancedShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (c advancedShow) Examples() []string {
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
	g, err := hx.GetGameName(hix)
	return g, nil, err
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
	return "Edit the current category."
}

func (advancedEdit) UsageArgs() string {
	return "<category...>"
}

func (c advancedEdit) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (c advancedEdit) Examples() []string {
	return []string{
		"minecraft",
		"just chatting",
	}
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

	hix, err := m.Here.IDExact()
	if err != nil {
		return "", nil, err
	}

	g, urr, err := hx.SetGame(hix, m.RawArgs(0))

	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}

	switch err {
	case nil:
		return fmt.Sprintf("Category set to: %s", g), nil, nil
	case twitch.ErrNoResults:
		return "Couldn't find the category, did you type the name correctly?", nil, nil
	default:
		return "", nil, err
	}
}
