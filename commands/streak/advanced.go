package streak

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/rs/zerolog/log"
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
	mod, err := m.Author.Mod()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if author is mod")
		return false
	}
	return mod
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
		AdvancedSet,
	}
}

func (advanced) Init() error {
	core.EventStreamOnlineHooks.Register(func(on *core.StreamOnline) {
		here, err := on.Here.ScopeLogical()
		if err != nil {
			log.Error().Err(err).Msg("failed to get place scope")
			return
		}
		err = Online(here, on.When)
		if err != nil {
			log.Error().Err(err).Msg("failed to save stream's online status")
			return
		}
	})

	core.EventStreamOfflineHooks.Register(func(off *core.StreamOffline) {
		here, err := off.Here.ScopeLogical()
		if err != nil {
			log.Error().Err(err).Msg("failed to get place scope")
			return
		}
		err = Offline(here, off.When)
		if err != nil {
			log.Error().Err(err).Msg("failed to save stream's offline status")
			return
		}
	})

	core.EventRedeemClaimHooks.Register(func(r *core.RedeemClaim) {
		person, err := r.Author.Scope()
		if err != nil {
			log.Error().Err(err).Msg("failed to get author scope")
			return
		}

		place, err := r.Here.ScopeLogical()
		if err != nil {
			log.Error().Err(err).Msg("failed to get place scope")
			return
		}

		id, err := EventIDGet(place)
		if err != nil {
			log.Error().Err(err).Msg("failed to get event id")
			return
		}

		if r.ID != id.String() {
			log.Debug().
				Str("got", r.ID).
				Str("expected", id.String()).
				Msg("redeem id doesn't match")
			return
		}

		if err := Appearance(person, place); err != nil {
			log.Error().Err(err).Msg("failed to update user streak")
		}
	})
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
	h, err := twitch.Frontend.Helix()
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
	h, err := twitch.Frontend.Helix()
	if err != nil {
		return err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}

	return Off(h, here)
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
	return "Set the ID of the."
}

func (advancedSet) UsageArgs() string {
	return "<id>"
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

func (c advancedSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return "Set the streak redeem.", nil, nil
}

func (advancedSet) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return EventIDSet(here, m.Command.Args[0])
}
