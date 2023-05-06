package streak

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	if m.Frontend.Type() != twitch.Type {
		return false
	}
	return m.Author.Mod()
}

func (advanced) Names() []string {
	return []string{
		"streak",
	}
}

func (advanced) Description() string {
	return "Control tracking of streaks."
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
		AdvancedOn,
		AdvancedOff,
	}
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (resp any, usrErr error, err error) {
	return m.Usage(), nil, nil
}

////////
//    //
// on //
//    //
////////

var AdvancedOn = advancedOn{}

type advancedOn struct{}

func (c advancedOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedOn) Names() []string {
	return core.AliasesOn
}

func (advancedOn) Description() string {
	return "Turn streak tracking on."
}

func (advancedOn) UsageArgs() string {
	return ""
}

func (c advancedOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedOn) Examples() []string {
	return nil
}

func (advancedOn) Parent() core.CommandStatic {
	return Advanced
}

func (advancedOn) Children() core.CommandsStatic {
	return nil
}

func (advancedOn) Init() error {
	return nil
}

func (c advancedOn) Run(m *core.Message) (resp any, usrErr error, err error) {
	if err := c.core(m); err != nil {
		return nil, nil, err
	}
	return "Streak tracking has been turned on.", nil, nil
}

func (advancedOn) core(m *core.Message) error {
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}

	return On(h, here, m.Here.ID())
}

/////////
//     //
// off //
//     //
/////////

var AdvancedOff = advancedOff{}

type advancedOff struct{}

func (c advancedOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedOff) Names() []string {
	return core.AliasesOff
}

func (advancedOff) Description() string {
	return "Turn streak tracking off."
}

func (advancedOff) UsageArgs() string {
	return ""
}

func (c advancedOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedOff) Examples() []string {
	return nil
}

func (advancedOff) Parent() core.CommandStatic {
	return Advanced
}

func (advancedOff) Children() core.CommandsStatic {
	return nil
}

func (advancedOff) Init() error {
	return nil
}

func (c advancedOff) Run(m *core.Message) (resp any, usrErr error, err error) {
	if err := c.core(m); err != nil {
		return nil, nil, err
	}
	return "Streak tracking has been turned off.", nil, nil
}

func (advancedOff) core(m *core.Message) error {
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}

	return Off(h, here)
}
