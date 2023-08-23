package god

import (
	"fmt"
	"strings"
	"sync"
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
		AdvancedAuto,
		AdvancedRedeem,
		AdvancedPersonality,
	}
}

func (advanced) Init() error {
	var mu sync.Mutex

	core.EventMessageHooks.Register(func(m *core.Message) {
		// Due to the fact that Talk can take a couple of seconds to return,
		// multiple auto-replies can be queued during that period, since
		// cmd_god_reply_last gets updated after the call.
		// To fix this, a mutex is used.
		mu.Lock()
		defer mu.Unlock()

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

		// Don't remember conversation as it is meant to be a random response,
		// not a discussion
		resp, err := Talk(-1, here, m.Raw)
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
		author, err := rc.Author.Scope()
		if err != nil {
			log.Error().Err(err).Msg("failed to get author scope")
			return
		}
		here, err := rc.Here.ScopeLogical()
		if err != nil {
			log.Error().Err(err).Msg("failed to get logical here")
			return
		}

		slog := log.With().
			Int64("person", author).
			Int64("place", here).
			Logger()

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

		resp, err := Talk(author, here, rc.Input)
		if err != nil {
			slog.Error().Err(err).Msg("failed to get gpt response")
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
	author, err := m.Author.Scope()
	if err != nil {
		return "", err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", err
	}
	return Talk(author, here, m.RawArgs(0))
}

//////////
//      //
// auto //
//      //
//////////

var AdvancedAuto = advancedAuto{}

type advancedAuto struct{}

func (c advancedAuto) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAuto) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (c advancedAuto) Names() []string {
	return []string{
		"auto",
	}
}

func (advancedAuto) Description() string {
	return "Control God's auto-replying."
}

func (c advancedAuto) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedAuto) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAuto) Examples() []string {
	return nil
}

func (advancedAuto) Parent() core.CommandStatic {
	return Advanced
}

func (advancedAuto) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedAutoReply,
		AdvancedAutoInterval,
	}
}

func (advancedAuto) Init() error {
	return nil
}

func (advancedAuto) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

////////////////
//            //
// auto reply //
//            //
////////////////

var AdvancedAutoReply = advancedAutoReply{}

type advancedAutoReply struct{}

func (c advancedAutoReply) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoReply) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoReply) Names() []string {
	return []string{
		"reply",
	}
}

func (advancedAutoReply) Description() string {
	return "Auto-replying related commands."
}

func (c advancedAutoReply) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedAutoReply) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoReply) Examples() []string {
	return nil
}

func (advancedAutoReply) Parent() core.CommandStatic {
	return AdvancedAuto
}

func (advancedAutoReply) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedAutoReplyShow,
		AdvancedAutoReplyOn,
		AdvancedAutoReplyOff,
	}
}

func (advancedAutoReply) Init() error {
	return nil
}

func (c advancedAutoReply) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

/////////////////////
//                 //
// auto reply show //
//                 //
/////////////////////

var AdvancedAutoReplyShow = advancedAutoReplyShow{}

type advancedAutoReplyShow struct{}

func (c advancedAutoReplyShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoReplyShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoReplyShow) Names() []string {
	return core.AliasesShow
}

func (advancedAutoReplyShow) Description() string {
	return "Show if auto-replying is on or off."
}

func (advancedAutoReplyShow) UsageArgs() string {
	return ""
}

func (c advancedAutoReplyShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoReplyShow) Examples() []string {
	return nil
}

func (advancedAutoReplyShow) Parent() core.CommandStatic {
	return AdvancedAutoReply
}

func (advancedAutoReplyShow) Children() core.CommandsStatic {
	return nil
}

func (advancedAutoReplyShow) Init() error {
	return nil
}

func (c advancedAutoReplyShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedAutoReplyShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	on, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(on),
	}
	return embed, nil, nil
}

func (c advancedAutoReplyShow) text(m *core.Message) (string, core.Urr, error) {
	on, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(on), nil, nil
}

func (advancedAutoReplyShow) fmt(on bool) string {
	if on {
		return "Auto-replying is on."
	}
	return "Auto-replying is off."
}

func (advancedAutoReplyShow) core(m *core.Message) (bool, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return false, err
	}
	return ReplyOnGet(here)
}

///////////////////
//               //
// auto reply on //
//               //
///////////////////

var AdvancedAutoReplyOn = advancedAutoReplyOn{}

type advancedAutoReplyOn struct{}

func (c advancedAutoReplyOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoReplyOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoReplyOn) Names() []string {
	return core.AliasesOn
}

func (advancedAutoReplyOn) Description() string {
	return "Turn auto-replying on."
}

func (advancedAutoReplyOn) UsageArgs() string {
	return ""
}

func (c advancedAutoReplyOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoReplyOn) Examples() []string {
	return nil
}

func (advancedAutoReplyOn) Parent() core.CommandStatic {
	return AdvancedAutoReply
}

func (advancedAutoReplyOn) Children() core.CommandsStatic {
	return nil
}

func (advancedAutoReplyOn) Init() error {
	return nil
}

func (c advancedAutoReplyOn) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedAutoReplyOn) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedAutoReplyOn) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedAutoReplyOn) fmt() string {
	return "Auto-replying has been turned on."
}

func (advancedAutoReplyOn) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return ReplyOnSet(here, true)
}

////////////////////
//                //
// auto reply off //
//                //
////////////////////

var AdvancedAutoReplyOff = advancedAutoReplyOff{}

type advancedAutoReplyOff struct{}

func (c advancedAutoReplyOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoReplyOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoReplyOff) Names() []string {
	return core.AliasesOff
}

func (advancedAutoReplyOff) Description() string {
	return "Turn auto-replying off."
}

func (advancedAutoReplyOff) UsageArgs() string {
	return ""
}

func (c advancedAutoReplyOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoReplyOff) Examples() []string {
	return nil
}

func (advancedAutoReplyOff) Parent() core.CommandStatic {
	return AdvancedAutoReply
}

func (advancedAutoReplyOff) Children() core.CommandsStatic {
	return nil
}

func (advancedAutoReplyOff) Init() error {
	return nil
}

func (c advancedAutoReplyOff) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedAutoReplyOff) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedAutoReplyOff) text(m *core.Message) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedAutoReplyOff) fmt() string {
	return "Auto-replying has been turned off."
}

func (advancedAutoReplyOff) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return ReplyOnSet(here, false)
}

///////////////////
//               //
// auto interval //
//               //
///////////////////

var AdvancedAutoInterval = advancedAutoInterval{}

type advancedAutoInterval struct{}

func (c advancedAutoInterval) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoInterval) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoInterval) Names() []string {
	return []string{
		"interval",
	}
}

func (advancedAutoInterval) Description() string {
	return "Control the interval between the auto-replies."
}

func (c advancedAutoInterval) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedAutoInterval) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoInterval) Examples() []string {
	return nil
}

func (advancedAutoInterval) Parent() core.CommandStatic {
	return AdvancedAuto
}

func (advancedAutoInterval) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedAutoIntervalShow,
		AdvancedAutoIntervalSet,
	}
}

func (advancedAutoInterval) Init() error {
	return nil
}

func (advancedAutoInterval) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

////////////////////////
//                    //
// auto interval show //
//                    //
////////////////////////

var AdvancedAutoIntervalShow = advancedAutoIntervalShow{}

type advancedAutoIntervalShow struct{}

func (c advancedAutoIntervalShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoIntervalShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoIntervalShow) Names() []string {
	return core.AliasesShow
}

func (advancedAutoIntervalShow) Description() string {
	return "Show the currently-set interval between the auto-replies."
}

func (c advancedAutoIntervalShow) UsageArgs() string {
	return ""
}

func (c advancedAutoIntervalShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoIntervalShow) Examples() []string {
	return nil
}

func (advancedAutoIntervalShow) Parent() core.CommandStatic {
	return AdvancedAutoInterval
}

func (advancedAutoIntervalShow) Children() core.CommandsStatic {
	return nil
}

func (advancedAutoIntervalShow) Init() error {
	return nil
}

func (c advancedAutoIntervalShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedAutoIntervalShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	interval, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval),
	}
	return embed, nil, nil
}

func (c advancedAutoIntervalShow) text(m *core.Message) (string, core.Urr, error) {
	interval, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval), nil, nil
}

func (advancedAutoIntervalShow) fmt(interval time.Duration) string {
	return "God will automatically reply once every " + interval.String()
}

func (advancedAutoIntervalShow) core(m *core.Message) (time.Duration, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Second, err
	}
	return ReplyIntervalGet(here)
}

///////////////////////
//                   //
// auto interval set //
//                   //
///////////////////////

var AdvancedAutoIntervalSet = advancedAutoIntervalSet{}

type advancedAutoIntervalSet struct{}

func (c advancedAutoIntervalSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAutoIntervalSet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAutoIntervalSet) Names() []string {
	return core.AliasesSet
}

func (advancedAutoIntervalSet) Description() string {
	return "Set the interval between the auto-replies."
}

func (c advancedAutoIntervalSet) UsageArgs() string {
	return "<seconds>"
}

func (c advancedAutoIntervalSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAutoIntervalSet) Examples() []string {
	return nil
}

func (advancedAutoIntervalSet) Parent() core.CommandStatic {
	return AdvancedAutoInterval
}

func (advancedAutoIntervalSet) Children() core.CommandsStatic {
	return nil
}

func (advancedAutoIntervalSet) Init() error {
	return nil
}

func (c advancedAutoIntervalSet) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedAutoIntervalSet) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	interval, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval, urr),
	}
	return embed, urr, nil
}

func (c advancedAutoIntervalSet) text(m *core.Message) (string, core.Urr, error) {
	interval, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval, urr), urr, nil
}

func (advancedAutoIntervalSet) fmt(interval time.Duration, urr core.Urr) string {
	switch urr {
	case nil:
		return fmt.Sprintf("Updated the interval to %s.", interval)
	case UrrIntervalTooShort:
		return fmt.Sprintf("The interval %s is too short, must be longer or equal to %s.", interval, core.MinGodInterval)
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedAutoIntervalSet) core(m *core.Message) (time.Duration, core.Urr, error) {
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

/////////////////
//             //
// personality //
//             //
/////////////////

var AdvancedPersonality = advancedPersonality{}

type advancedPersonality struct{}

func (c advancedPersonality) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonality) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonality) Names() []string {
	return []string{
		"personality",
		"mood",
		"cosplay",
	}
}

func (advancedPersonality) Description() string {
	return "Control God's personality."
}

func (c advancedPersonality) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedPersonality) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonality) Examples() []string {
	return nil
}

func (advancedPersonality) Parent() core.CommandStatic {
	return Advanced
}

func (advancedPersonality) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedPersonalityShow,
		AdvancedPersonalitySet,
		AdvancedPersonalityAdd,
		AdvancedPersonalityEdit,
		AdvancedPersonalityDelete,
		AdvancedPersonalityInfo,
		AdvancedPersonalityList,
	}
}

func (advancedPersonality) Init() error {
	return nil
}

func (advancedPersonality) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

//////////////////////
//                  //
// personality show //
//                  //
//////////////////////

var AdvancedPersonalityShow = advancedPersonalityShow{}

type advancedPersonalityShow struct{}

func (c advancedPersonalityShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityShow) Names() []string {
	return core.AliasesShow
}

func (advancedPersonalityShow) Description() string {
	return "Show God's current personality."
}

func (advancedPersonalityShow) UsageArgs() string {
	return ""
}

func (c advancedPersonalityShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityShow) Examples() []string {
	return nil
}

func (advancedPersonalityShow) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityShow) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityShow) Init() error {
	return nil
}

func (c advancedPersonalityShow) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPersonalityShow) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	personality, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(personality),
	}
	return embed, nil, nil
}

func (c advancedPersonalityShow) text(m *core.Message) (string, core.Urr, error) {
	personality, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(personality), nil, nil
}

func (advancedPersonalityShow) fmt(p Personality) string {
	return "Current personality is: " + p.Name
}

func (advancedPersonalityShow) core(m *core.Message) (Personality, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return Personality{}, err
	}
	p, err := PersonalityActive(here)
	if err != nil {
		return Personality{}, err
	}
	return p, nil
}

/////////////////////
//                 //
// personality set //
//                 //
/////////////////////

var AdvancedPersonalitySet = advancedPersonalitySet{}

type advancedPersonalitySet struct{}

func (c advancedPersonalitySet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalitySet) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalitySet) Names() []string {
	return core.AliasesSet
}

func (advancedPersonalitySet) Description() string {
	return "Set God's personality."
}

func (c advancedPersonalitySet) UsageArgs() string {
	return "<name>"
}

func (c advancedPersonalitySet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalitySet) Examples() []string {
	return nil
}

func (advancedPersonalitySet) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (c advancedPersonalitySet) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalitySet) Init() error {
	return nil
}

func (c advancedPersonalitySet) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedPersonalitySet) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	name, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt("**"+name+"**", urr),
	}
	return embed, urr, nil
}

func (c advancedPersonalitySet) text(m *core.Message) (string, core.Urr, error) {
	name, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(name, urr), urr, nil
}

func (advancedPersonalitySet) fmt(name string, urr core.Urr) string {
	switch urr {
	case nil:
		return "Set personality to " + name
	case UrrPersonalityNotFound:
		return "Couldn't find " + name + " personality."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedPersonalitySet) core(m *core.Message) (string, core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", nil, err
	}
	return PersonalitySet(here, m.Command.Args[0])
}

/////////////////////
//                 //
// personality add //
//                 //
/////////////////////

var AdvancedPersonalityAdd = advancedPersonalityAdd{}

type advancedPersonalityAdd struct{}

func (c advancedPersonalityAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityAdd) Names() []string {
	return core.AliasesAdd
}

func (advancedPersonalityAdd) Description() string {
	return "Add a new God personality."
}

func (advancedPersonalityAdd) UsageArgs() string {
	return "<personality> <prompt...>"
}

func (c advancedPersonalityAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityAdd) Examples() []string {
	return nil
}

func (advancedPersonalityAdd) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityAdd) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityAdd) Init() error {
	return nil
}

func (c advancedPersonalityAdd) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPersonalityAdd) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedPersonalityAdd) text(m *core.Message) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedPersonalityAdd) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Added personality."
	default:
		return urr.Error()
	}
}

func (advancedPersonalityAdd) core(m *core.Message) (core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return PersonalityAdd(here, m.Command.Args[0], m.RawArgs(1))
}

//////////////////////
//                  //
// personality edit //
//                  //
//////////////////////

var AdvancedPersonalityEdit = advancedPersonalityEdit{}

type advancedPersonalityEdit struct{}

func (c advancedPersonalityEdit) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityEdit) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityEdit) Names() []string {
	return core.AliasesEdit
}

func (advancedPersonalityEdit) Description() string {
	return "Edit one of God's personalities."
}

func (advancedPersonalityEdit) UsageArgs() string {
	return "<personality> <prompt...>"
}

func (c advancedPersonalityEdit) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityEdit) Examples() []string {
	return nil
}

func (advancedPersonalityEdit) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityEdit) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityEdit) Init() error {
	return nil
}

func (c advancedPersonalityEdit) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPersonalityEdit) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedPersonalityEdit) text(m *core.Message) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedPersonalityEdit) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Edited personality."
	default:
		return urr.Error()
	}
}

func (advancedPersonalityEdit) core(m *core.Message) (core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return PersonalityEdit(here, m.Command.Args[0], m.RawArgs(1))
}

////////////////////////
//                    //
// personality delete //
//                    //
////////////////////////

var AdvancedPersonalityDelete = advancedPersonalityDelete{}

type advancedPersonalityDelete struct{}

func (c advancedPersonalityDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedPersonalityDelete) Description() string {
	return "Delete one of God's personalities."
}

func (advancedPersonalityDelete) UsageArgs() string {
	return "<personality>"
}

func (c advancedPersonalityDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityDelete) Examples() []string {
	return nil
}

func (advancedPersonalityDelete) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityDelete) Init() error {
	return nil
}

func (c advancedPersonalityDelete) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedPersonalityDelete) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedPersonalityDelete) text(m *core.Message) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedPersonalityDelete) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Deleted personality."
	default:
		return urr.Error()
	}
}

func (advancedPersonalityDelete) core(m *core.Message) (core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return PersonalityDelete(here, m.Command.Args[0])
}

//////////////////////
//                  //
// personality info //
//                  //
//////////////////////

var AdvancedPersonalityInfo = advancedPersonalityInfo{}

type advancedPersonalityInfo struct{}

func (c advancedPersonalityInfo) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityInfo) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityInfo) Names() []string {
	return []string{
		"info",
	}
}

func (advancedPersonalityInfo) Description() string {
	return "View information on the specified personality."
}

func (advancedPersonalityInfo) UsageArgs() string {
	return "<personality>"
}

func (c advancedPersonalityInfo) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityInfo) Examples() []string {
	return nil
}

func (advancedPersonalityInfo) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityInfo) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityInfo) Init() error {
	return nil
}

func (c advancedPersonalityInfo) Run(m *core.Message) (any, core.Urr, error) {
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

func (c advancedPersonalityInfo) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	personality, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if urr != nil {
		embed := &dg.MessageEmbed{
			Description: urr.Error(),
		}
		return embed, urr, nil
	}
	embed := &dg.MessageEmbed{
		Title:       "Personality: " + personality.Name,
		Description: "**Prompt**\n" + personality.Prompt,
	}
	return embed, nil, nil
}

func (c advancedPersonalityInfo) text(m *core.Message) (string, core.Urr, error) {
	personality, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if urr != nil {
		return urr.Error(), urr, nil
	}
	return personality.Name + ": " + personality.Prompt, nil, nil
}

func (advancedPersonalityInfo) core(m *core.Message) (Personality, core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return Personality{}, nil, err
	}
	return PersonalityGet(here, m.Command.Args[0])
}

//////////////////////
//                  //
// personality list //
//                  //
//////////////////////

var AdvancedPersonalityList = advancedPersonalityList{}

type advancedPersonalityList struct{}

func (c advancedPersonalityList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPersonalityList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPersonalityList) Names() []string {
	return core.AliasesList
}

func (advancedPersonalityList) Description() string {
	return "List all the available personalities."
}

func (advancedPersonalityList) UsageArgs() string {
	return ""
}

func (c advancedPersonalityList) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPersonalityList) Examples() []string {
	return nil
}

func (advancedPersonalityList) Parent() core.CommandStatic {
	return AdvancedPersonality
}

func (advancedPersonalityList) Children() core.CommandsStatic {
	return nil
}

func (advancedPersonalityList) Init() error {
	return nil
}

func (c advancedPersonalityList) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPersonalityList) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	active, personalities, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	var b strings.Builder
	for _, p := range personalities {
		if p.Name == active.Name {
			p.Name = "**" + p.Name + "**"
		}
		b.WriteString("- ")
		b.WriteString(p.Name)
		b.WriteString("\n")
	}
	embed := &dg.MessageEmbed{
		Description: b.String(),
	}
	return embed, nil, nil
}

func (c advancedPersonalityList) text(m *core.Message) (string, core.Urr, error) {
	_, personalities, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	var names []string
	for _, p := range personalities {
		names = append(names, p.Name)
	}
	return "Personalities: " + strings.Join(names, ", "), nil, nil
}

func (advancedPersonalityList) core(m *core.Message) (Personality, []Personality, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return Personality{}, nil, err
	}
	ps, err := PersonalitiesList(here)
	if err != nil {
		return Personality{}, nil, err
	}
	active, err := PersonalityActive(here)
	if err != nil {
		return Personality{}, nil, err
	}
	return active, ps, nil
}
