package title

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	if m.Frontend != frontends.Twitch {
		return false
	}
	return m.Mod()
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
	return "Show the current title."
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
	case frontends.Twitch:
		return c.twitch(m)
	default:
		panic("this should never happen")
	}
}

func (advancedShow) twitch(m *core.Message) (string, error, error) {
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}

	t, err := h.GetTitle(m.Channel.ID)
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

func (c advancedEdit) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedEdit) Names() []string {
	return core.Edit
}

func (advancedEdit) Description() string {
	return "Edit the current title."
}

func (advancedEdit) UsageArgs() string {
	return "<title...>"
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

func (c advancedEdit) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Twitch:
		return c.twitch(m)
	default:
		panic("this should never happen")
	}
}

func (advancedEdit) twitch(m *core.Message) (string, error, error) {
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}

	title := m.RawArgs(0)

	usrErr, err := h.SetTitle(m.Channel.ID, title)
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("Title set to: %s", title), nil, nil
}