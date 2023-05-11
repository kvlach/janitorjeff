package streak

import (
	"fmt"
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/google/uuid"
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
		AdvancedShow,
		AdvancedRedeem,
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

		id, usrErr, err := RedeemGet(place)
		if usrErr != nil || err != nil {
			log.Error().Err(err).Interface("usrErr", usrErr).Msg("failed to get event id")
			return
		}

		if r.ID != id.String() {
			log.Debug().
				Str("got", r.ID).
				Str("expected", id.String()).
				Msg("redeem id doesn't match")
			return
		}

		streak, err := Appearance(person, place)
		if err != nil {
			log.Error().Err(err).Msg("failed to update user streak")
			return
		}

		m, err := core.Frontends.CreateMessage(person, place, "")
		if err != nil {
			log.Error().Err(err).Msg("failed to create message object")
			return
		}

		display, err := r.Author.DisplayName()
		if err != nil {
			log.Error().Err(err).Msg("failed to get display name")
			return
		}

		var times string
		if streak == 1 {
			times = "time"
		} else {
			times = "times"
		}

		resp := fmt.Sprintf("%s has paid their taxes %d %s in a row!", display, streak, times)
		if _, err := m.Client.Send(resp, nil); err != nil {
			log.Error().Err(err).Msg("failed to send streak message")
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
	return core.AliasesShow
}

func (advancedShow) Description() string {
	return "Show the current streak."
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

func (c advancedShow) Run(m *core.Message) (any, error, error) {
	streak, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return fmt.Sprintf("Current streak is: %d", streak), nil, nil
}

func (advancedShow) core(m *core.Message) (int64, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return 0, err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return 0, err
	}
	return Get(author, here)
}

////////////
//        //
// redeem //
//        //
////////////

var AdvancedRedeem = advancedRedeem{}

type advancedRedeem struct{}

func (c advancedRedeem) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRedeem) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedRedeem) Names() []string {
	return []string{
		"redeem",
	}
}

func (advancedRedeem) Description() string {
	return "Control which redeem triggers the streak."
}

func (c advancedRedeem) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedRedeem) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRedeem) Examples() []string {
	return nil
}

func (advancedRedeem) Parent() core.CommandStatic {
	return Advanced
}

func (advancedRedeem) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedRedeemShow,
		AdvancedRedeemSet,
	}
}

func (advancedRedeem) Init() error {
	return nil
}

func (advancedRedeem) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////////////
//             //
// redeem show //
//             //
/////////////////

var AdvancedRedeemShow = advancedRedeemShow{}

type advancedRedeemShow struct{}

func (c advancedRedeemShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRedeemShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedRedeemShow) Names() []string {
	return core.AliasesShow
}

func (advancedRedeemShow) Description() string {
	return "Show what the current redeem is set to."
}

func (advancedRedeemShow) UsageArgs() string {
	return ""
}

func (c advancedRedeemShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRedeemShow) Examples() []string {
	return nil
}

func (advancedRedeemShow) Parent() core.CommandStatic {
	return AdvancedRedeem
}

func (advancedRedeemShow) Children() core.CommandsStatic {
	return nil
}

func (advancedRedeemShow) Init() error {
	return nil
}

func (c advancedRedeemShow) Run(m *core.Message) (any, error, error) {
	u, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return c.fmt(u, usrErr), usrErr, nil
}

func (advancedRedeemShow) fmt(u uuid.UUID, usrErr error) string {
	switch usrErr {
	case nil:
		return "The streak tracking redeem is set to: " + u.String()
	case ErrRedeemNotSet:
		return "The streak tracking redeem has not been set."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedRedeemShow) core(m *core.Message) (uuid.UUID, error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	return RedeemGet(here)
}

////////////////
//            //
// redeem set //
//            //
////////////////

var AdvancedRedeemSet = advancedRedeemSet{}

type advancedRedeemSet struct{}

func (c advancedRedeemSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRedeemSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedRedeemSet) Names() []string {
	return []string{
		"set",
	}
}

func (advancedRedeemSet) Description() string {
	return "Set the ID of the."
}

func (advancedRedeemSet) UsageArgs() string {
	return "<id>"
}

func (c advancedRedeemSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRedeemSet) Examples() []string {
	return nil
}

func (advancedRedeemSet) Parent() core.CommandStatic {
	return AdvancedRedeem
}

func (advancedRedeemSet) Children() core.CommandsStatic {
	return nil
}

func (advancedRedeemSet) Init() error {
	return nil
}

func (c advancedRedeemSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return "Set the streak redeem.", nil, nil
}

func (advancedRedeemSet) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return RedeemSet(here, m.Command.Args[0])
}
