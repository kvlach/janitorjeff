package god

import (
	"fmt"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var UrrInvalidInterval = core.UrrNew("Expected an integer number as the interval.")

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	mod, err := m.Author.Moderator()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if author is mod")
		return false
	}
	return mod
}

func (advanced) Names() []string {
	return []string{
		"god",
	}
}

func (advanced) Description() string {
	return "Control God."
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
		AdvancedTalk,
		AdvancedReply,
		AdvancedInterval,
		AdvancedRedeem,
		AdvancedMood,
	}
}

func (advanced) Init() error {
	core.EventMessageHooks.Register(func(m *core.Message) {
		here, err := m.Here.ScopeLogical()
		if err != nil {
			return
		}

		tx, err := core.DB.Begin()
		if err != nil {
			log.Error().Err(err).Msg("failed to begin transaction")
			return
		}
		//goland:noinspection GoUnhandledErrorResult
		defer tx.Rollback()

		// TODO: If the place settings have just been generated and added to
		// the cache then we need a way to remove them from the cache in case
		// the transaction gets rolled back. Otherwise, we end up in a situation
		// where the cache thinks that the settings have already been generated
		// but they in fact haven't, since the transaction got rolled back.

		on, err := tx.PlaceGet("cmd_god_reply_on", here).Bool()
		if err != nil {
			log.Error().Err(err).Msg("reply not on, skipping")
			return
		}
		if !on {
			// Must commit in case the place settings haven't been generated before
			tx.Commit()
			return
		}

		interval, err := tx.PlaceGet("cmd_god_reply_interval", here).Duration()
		if err != nil {
			return
		}
		last, err := tx.PlaceGet("cmd_god_reply_last", here).Time()
		if err != nil {
			return
		}
		diff := time.Now().UTC().Sub(last)
		if interval > diff {
			log.Debug().
				Interface("interval", interval).
				Interface("diff", diff).
				Msg("shouldn't reply yet, skipping")
			tx.Commit()
			return
		}

		resp, err := Talk(m.Raw, here)
		if err != nil {
			log.Debug().Err(err).Msg("failed to communicate with god")
			return
		}

		if _, err := m.Client.Natural(resp, nil); err != nil {
			log.Error().Err(err).Msg("failed to send message")
			return
		}

		if err := tx.PlaceSet("cmd_god_reply_last", here, time.Now().UTC().Unix()); err != nil {
			log.Debug().Err(err).Msg("error while trying to set reply")
			return
		}

		if err := tx.Commit(); err != nil {
			log.Error().Err(err).Msg("couldn't commit")
		}
	})

	core.EventRedeemClaimHooks.Register(func(rc *core.RedeemClaim) {
		here, err := rc.Here.ScopeLogical()
		if err != nil {
			log.Error().Err(err).Msg("failed to get logical here")
		}

		slog := log.With().Int64("place", here).Logger()

		rid, urr, err := RedeemGet(here)
		if err != nil {
			slog.Error().Err(err).Msg("failed to get redeem id")
			return
		}
		if urr != nil {
			slog.Error().Msg(AdvancedRedeemShow.fmt(rid, urr))
			return
		}

		if rc.ID != rid.String() {
			slog.Debug().
				Str("got", rc.ID).
				Str("expected", rid.String()).
				Msg("not the expected redeem")
			return
		}

		resp, err := Talk(rc.Input, here)
		if err != nil {
			slog.Error().Err(err).Msg("failed to get gpt response")
			return
		}

		author, err := rc.Author.Scope()
		if err != nil {
			log.Error().Err(err).Msg("failed to get author scope")
			return
		}

		slog = slog.With().Int64("person", author).Logger()

		m, err := core.Frontends.CreateMessage(author, here, "")
		if err != nil {
			slog.Error().Err(err).Msg("failed to create message")
			return
		}

		if _, err = m.Client.Natural(resp, nil); err != nil {
			slog.Error().Err(err).Msg("failed to write message")
			return
		}
	})

	return nil
}

func (c advanced) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////
//      //
// talk //
//      //
//////////

var AdvancedTalk = advancedTalk{}

type advancedTalk struct{}

func (c advancedTalk) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTalk) Permitted(m *core.Message) bool {
	return true
}

func (advancedTalk) Names() []string {
	return []string{
		"talk",
		"speak",
		"ask",
	}
}

func (advancedTalk) Description() string {
	return "Talk to God."
}

func (advancedTalk) UsageArgs() string {
	return "<text>"
}

func (c advancedTalk) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTalk) Examples() []string {
	return nil
}

func (advancedTalk) Parent() core.CommandStatic {
	return Advanced
}

func (advancedTalk) Children() core.CommandsStatic {
	return nil
}

func (advancedTalk) Init() error {
	return nil
}

func (c advancedTalk) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTalk) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	resp, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Title:       "God says...",
		Description: resp,
	}
	return embed, nil, nil
}

func (c advancedTalk) text(m *core.Message) (string, core.Urr, error) {
	resp, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return resp, nil, nil
}

func (advancedTalk) core(m *core.Message) (string, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", err
	}
	return Talk(m.RawArgs(0), here)
}

///////////
//       //
// reply //
//       //
///////////

var AdvancedReply = advancedReply{}

type advancedReply struct{}

func (c advancedReply) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReply) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReply) Names() []string {
	return []string{
		"reply",
	}
}

func (advancedReply) Description() string {
	return "Auto-replying related commands."
}

func (c advancedReply) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedReply) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedReply) Examples() []string {
	return nil
}

func (advancedReply) Parent() core.CommandStatic {
	return Advanced
}

func (advancedReply) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedReplyShow,
		AdvancedReplyOn,
		AdvancedReplyOff,
	}
}

func (advancedReply) Init() error {
	return nil
}

func (c advancedReply) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

////////////////
//            //
// reply show //
//            //
////////////////

var AdvancedReplyShow = advancedReplyShow{}

type advancedReplyShow struct{}

func (c advancedReplyShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReplyShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReplyShow) Names() []string {
	return core.AliasesShow
}

func (advancedReplyShow) Description() string {
	return "Show if auto-replying is on or off."
}

func (advancedReplyShow) UsageArgs() string {
	return ""
}

func (c advancedReplyShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedReplyShow) Examples() []string {
	return nil
}

func (advancedReplyShow) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyShow) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyShow) Init() error {
	return nil
}

func (c advancedReplyShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	on, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(on),
	}
	return embed, nil, nil
}

func (c advancedReplyShow) text(m *core.Message) (string, core.Urr, error) {
	on, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(on), nil, nil
}

func (advancedReplyShow) fmt(on bool) string {
	if on {
		return "Auto-replying is on."
	}
	return "Auto-replying is off."
}

func (advancedReplyShow) core(m *core.Message) (bool, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return false, err
	}
	return ReplyOnGet(here)
}

//////////////
//          //
// reply on //
//          //
//////////////

var AdvancedReplyOn = advancedReplyOn{}

type advancedReplyOn struct{}

func (c advancedReplyOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReplyOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReplyOn) Names() []string {
	return core.AliasesOn
}

func (advancedReplyOn) Description() string {
	return "Turn auto-replying on."
}

func (advancedReplyOn) UsageArgs() string {
	return ""
}

func (c advancedReplyOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedReplyOn) Examples() []string {
	return nil
}

func (advancedReplyOn) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyOn) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyOn) Init() error {
	return nil
}

func (c advancedReplyOn) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyOn) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedReplyOn) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedReplyOn) fmt() string {
	return "Auto-replying has been turned on."
}

func (advancedReplyOn) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return ReplyOnSet(here, true)
}

///////////////
//           //
// reply off //
//           //
///////////////

var AdvancedReplyOff = advancedReplyOff{}

type advancedReplyOff struct{}

func (c advancedReplyOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReplyOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReplyOff) Names() []string {
	return core.AliasesOff
}

func (advancedReplyOff) Description() string {
	return "Turn auto-replying off."
}

func (advancedReplyOff) UsageArgs() string {
	return ""
}

func (c advancedReplyOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedReplyOff) Examples() []string {
	return nil
}

func (advancedReplyOff) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyOff) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyOff) Init() error {
	return nil
}

func (c advancedReplyOff) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyOff) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedReplyOff) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedReplyOff) fmt() string {
	return "Auto-replying has been turned off."
}

func (advancedReplyOff) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return ReplyOnSet(here, false)
}

//////////////
//          //
// interval //
//          //
//////////////

var AdvancedInterval = advancedInterval{}

type advancedInterval struct{}

func (c advancedInterval) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedInterval) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedInterval) Names() []string {
	return []string{
		"interval",
	}
}

func (advancedInterval) Description() string {
	return "Control the interval between the auto-replies."
}

func (c advancedInterval) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedInterval) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedInterval) Examples() []string {
	return nil
}

func (advancedInterval) Parent() core.CommandStatic {
	return Advanced
}

func (advancedInterval) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedIntervalShow,
		AdvancedIntervalSet,
	}
}

func (advancedInterval) Init() error {
	return nil
}

func (advancedInterval) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

///////////////////
//               //
// interval show //
//               //
///////////////////

var AdvancedIntervalShow = advancedIntervalShow{}

type advancedIntervalShow struct{}

func (c advancedIntervalShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedIntervalShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedIntervalShow) Names() []string {
	return core.AliasesShow
}

func (advancedIntervalShow) Description() string {
	return "Show the currently-set interval between the auto-replies."
}

func (c advancedIntervalShow) UsageArgs() string {
	return ""
}

func (c advancedIntervalShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedIntervalShow) Examples() []string {
	return nil
}

func (advancedIntervalShow) Parent() core.CommandStatic {
	return AdvancedInterval
}

func (advancedIntervalShow) Children() core.CommandsStatic {
	return nil
}

func (advancedIntervalShow) Init() error {
	return nil
}

func (c advancedIntervalShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedIntervalShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	interval, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval),
	}
	return embed, nil, nil
}

func (c advancedIntervalShow) text(m *core.Message) (string, core.Urr, error) {
	interval, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval), nil, nil
}

func (advancedIntervalShow) fmt(interval time.Duration) string {
	return "God will automatically reply once every " + interval.String()
}

func (advancedIntervalShow) core(m *core.Message) (time.Duration, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Second, err
	}
	return ReplyIntervalGet(here)
}

//////////////////
//              //
// interval set //
//              //
//////////////////

var AdvancedIntervalSet = advancedIntervalSet{}

type advancedIntervalSet struct{}

func (c advancedIntervalSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedIntervalSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedIntervalSet) Names() []string {
	return core.AliasesSet
}

func (advancedIntervalSet) Description() string {
	return "Set the interval between the auto-replies."
}

func (c advancedIntervalSet) UsageArgs() string {
	return "<seconds>"
}

func (c advancedIntervalSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedIntervalSet) Examples() []string {
	return nil
}

func (advancedIntervalSet) Parent() core.CommandStatic {
	return AdvancedInterval
}

func (advancedIntervalSet) Children() core.CommandsStatic {
	return nil
}

func (advancedIntervalSet) Init() error {
	return nil
}

func (c advancedIntervalSet) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedIntervalSet) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	interval, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval, urr),
	}
	return embed, urr, nil
}

func (c advancedIntervalSet) text(m *core.Message) (string, core.Urr, error) {
	interval, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval, urr), urr, nil
}

func (advancedIntervalSet) fmt(interval time.Duration, urr core.Urr) string {
	switch urr {
	case nil:
		return fmt.Sprintf("Updated the interval to %s.", interval)
	case UrrIntervalTooShort:
		return fmt.Sprintf("The interval %s is too short, must be longer or equal to %s.", interval, core.MinGodInterval)
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedIntervalSet) core(m *core.Message) (time.Duration, core.Urr, error) {
	interval, err := time.ParseDuration(m.Command.Args[0])
	if err != nil {
		return 0, nil, err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Second, nil, err
	}
	urr, err := ReplyIntervalSet(here, interval)
	return interval, urr, err
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
	return "Control which redeem triggers god."
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

func (advancedRedeem) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c advancedRedeemShow) Run(m *core.Message) (any, core.Urr, error) {
	u, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return c.fmt(u, urr), urr, nil
}

func (advancedRedeemShow) fmt(u uuid.UUID, urr core.Urr) string {
	switch urr {
	case nil:
		return "The god redeem is set to: " + u.String()
	case core.UrrValNil:
		return "The god redeem has not been set."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedRedeemShow) core(m *core.Message) (uuid.UUID, core.Urr, error) {
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
	return core.AliasesSet
}

func (advancedRedeemSet) Description() string {
	return "Set the ID of the god redeem."
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

func (c advancedRedeemSet) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return "Set the god redeem.", nil, nil
}

func (advancedRedeemSet) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return RedeemSet(here, m.Command.Args[0])
}

//////////
//      //
// mood //
//      //
//////////

var AdvancedMood = advancedMood{}

type advancedMood struct{}

func (c advancedMood) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMood) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMood) Names() []string {
	return []string{
		"mood",
		"moods",
	}
}

func (advancedMood) Description() string {
	return "Control God's mood."
}

func (c advancedMood) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedMood) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMood) Examples() []string {
	return nil
}

func (advancedMood) Parent() core.CommandStatic {
	return Advanced
}

func (advancedMood) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedMoodShow,
		AdvancedMoodSet,
	}
}

func (advancedMood) Init() error {
	return nil
}

func (advancedMood) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

///////////////
//           //
// mood show //
//           //
///////////////

var AdvancedMoodShow = advancedMoodShow{}

type advancedMoodShow struct{}

func (c advancedMoodShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodShow) Names() []string {
	return core.AliasesShow
}

func (advancedMoodShow) Description() string {
	return "Show God's current mood."
}

func (advancedMoodShow) UsageArgs() string {
	return ""
}

func (c advancedMoodShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodShow) Examples() []string {
	return nil
}

func (advancedMoodShow) Parent() core.CommandStatic {
	return AdvancedMood
}

func (advancedMoodShow) Children() core.CommandsStatic {
	return nil
}

func (advancedMoodShow) Init() error {
	return nil
}

func (c advancedMoodShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return nil, nil, nil
	}
}

func (c advancedMoodShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	mood, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(mood),
	}
	return embed, nil, nil
}

func (advancedMoodShow) fmt(mood Mood) string {
	return mood.String()
}

func (advancedMoodShow) core(m *core.Message) (Mood, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return 0, err
	}
	mood, err := MoodGet(here)
	if err != nil {
		return 0, err
	}
	return mood, nil
}

//////////////
//          //
// mood set //
//          //
//////////////

var AdvancedMoodSet = advancedMoodSet{}

type advancedMoodSet struct{}

func (c advancedMoodSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodSet) Names() []string {
	return core.AliasesSet
}

func (advancedMoodSet) Description() string {
	return "Set God's mood."
}

func (c advancedMoodSet) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedMoodSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodSet) Examples() []string {
	return nil
}

func (advancedMoodSet) Parent() core.CommandStatic {
	return AdvancedMood
}

func (c advancedMoodSet) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedMoodSetDefault,
		AdvancedMoodSetRude,
		AdvancedMoodSetSad,
		AdvancedMoodSetNeutral,
	}
}

func (advancedMoodSet) Init() error {
	return nil
}

func (advancedMoodSet) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////////////////
//                  //
// mood set default //
//                  //
//////////////////////

var AdvancedMoodSetDefault = advancedMoodSetDefault{}

type advancedMoodSetDefault struct{}

func (c advancedMoodSetDefault) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodSetDefault) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodSetDefault) Names() []string {
	return []string{
		"default",
	}
}

func (advancedMoodSetDefault) Description() string {
	return "Revert to the default mood."
}

func (c advancedMoodSetDefault) UsageArgs() string {
	return ""
}

func (c advancedMoodSetDefault) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodSetDefault) Examples() []string {
	return nil
}

func (advancedMoodSetDefault) Parent() core.CommandStatic {
	return AdvancedMoodSet
}

func (advancedMoodSetDefault) Children() core.CommandsStatic {
	return nil
}

func (advancedMoodSetDefault) Init() error {
	return nil
}

func (c advancedMoodSetDefault) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedMoodSetDefault) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedMoodSetDefault) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedMoodSetDefault) fmt() string {
	return "Set the mood to default."
}

func (advancedMoodSetDefault) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return MoodSet(here, MoodDefault)
}

///////////////////
//               //
// mood set rude //
//               //
///////////////////

var AdvancedMoodSetRude = advancedMoodSetRude{}

type advancedMoodSetRude struct{}

func (c advancedMoodSetRude) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodSetRude) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodSetRude) Names() []string {
	return []string{
		"rude",
	}
}

func (advancedMoodSetRude) Description() string {
	return "Make God rude."
}

func (c advancedMoodSetRude) UsageArgs() string {
	return ""
}

func (c advancedMoodSetRude) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodSetRude) Examples() []string {
	return nil
}

func (advancedMoodSetRude) Parent() core.CommandStatic {
	return AdvancedMoodSet
}

func (advancedMoodSetRude) Children() core.CommandsStatic {
	return nil
}

func (advancedMoodSetRude) Init() error {
	return nil
}

func (c advancedMoodSetRude) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedMoodSetRude) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedMoodSetRude) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedMoodSetRude) fmt() string {
	return "God will now be very rude!"
}

func (advancedMoodSetRude) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return MoodSet(here, MoodRude)
}

//////////////////
//              //
// mood set sad //
//              //
//////////////////

var AdvancedMoodSetSad = advancedMoodSetSad{}

type advancedMoodSetSad struct{}

func (c advancedMoodSetSad) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodSetSad) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodSetSad) Names() []string {
	return []string{
		"sad",
	}
}

func (advancedMoodSetSad) Description() string {
	return "Make God sad :("
}

func (c advancedMoodSetSad) UsageArgs() string {
	return ""
}

func (c advancedMoodSetSad) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodSetSad) Examples() []string {
	return nil
}

func (advancedMoodSetSad) Parent() core.CommandStatic {
	return AdvancedMoodSet
}

func (advancedMoodSetSad) Children() core.CommandsStatic {
	return nil
}

func (advancedMoodSetSad) Init() error {
	return nil
}

func (c advancedMoodSetSad) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedMoodSetSad) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedMoodSetSad) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedMoodSetSad) fmt() string {
	return "God is now sad :("
}

func (advancedMoodSetSad) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return MoodSet(here, MoodSad)
}

//////////////////////
//                  //
// mood set neutral //
//                  //
//////////////////////

var AdvancedMoodSetNeutral = advancedMoodSetNeutral{}

type advancedMoodSetNeutral struct{}

func (c advancedMoodSetNeutral) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedMoodSetNeutral) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedMoodSetNeutral) Names() []string {
	return []string{
		"neutral",
	}
}

func (advancedMoodSetNeutral) Description() string {
	return "Make God respond a generally neutral tone."
}

func (c advancedMoodSetNeutral) UsageArgs() string {
	return ""
}

func (c advancedMoodSetNeutral) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedMoodSetNeutral) Examples() []string {
	return nil
}

func (advancedMoodSetNeutral) Parent() core.CommandStatic {
	return AdvancedMoodSet
}

func (advancedMoodSetNeutral) Children() core.CommandsStatic {
	return nil
}

func (advancedMoodSetNeutral) Init() error {
	return nil
}

func (c advancedMoodSetNeutral) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedMoodSetNeutral) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedMoodSetNeutral) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedMoodSetNeutral) fmt() string {
	return "Not really much personality here."
}

func (advancedMoodSetNeutral) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return MoodSet(here, MoodNeutral)
}
