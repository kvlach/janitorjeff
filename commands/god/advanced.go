package god

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"
	"github.com/janitorjeff/jeff-bot/frontends/twitch"
	"github.com/rs/zerolog/log"

	dg "github.com/bwmarrin/discordgo"
)

var ErrInvalidInterval = errors.New("Expected an integer number as the interval.")

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return m.Author.Mod()
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

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedTalk,
		AdvancedReply,
		AdvancedInterval,
	}
}

func (advanced) Init() error {
	core.Hooks.Register(func(m *core.Message) {
		here, err := m.Here.ScopeLogical()
		if err != nil {
			return
		}

		if on, err := ReplyOnGet(here); err != nil || !on {
			log.Debug().Err(err).Msg("reply not on, skipping")
			return
		}

		if should, err := ShouldReply(here); err != nil || !should {
			log.Debug().Err(err).Msg("shouldn't reply yet, skipping")
			return
		}

		resp, err := Talk(m.Raw)
		if err != nil {
			return
		}

		// Make it so that on twitch it sometimes mentions and other times
		// doesn't the person it's replying to. This can make it seem more
		// natural as opposed to just a dry response by the bot, which is also
		// why m.Write isn't used when we the person is mentioned, since we
		// want to avoid the arrow in the response (@person -> response). The
		// whole thing is a bit hacky, but what can you do, the people have
		// asked for this.
		if m.Frontend.Type() == twitch.Frontend.Type() {
			rand.Seed(time.Now().UnixNano())
			// need this to only happen 30% of the time
			if num := rand.Intn(10); num < 3 {
				resp = "@" + m.Author.DisplayName() + " " + resp
			}
			m.Client.Send(resp, nil)
		} else {
			m.Write(resp, nil)
		}

		if err := ReplyLastSet(here, time.Now()); err != nil {
			log.Debug().Err(err).Msg("error while trying to set reply")
			return
		}
	})
	return nil
}

func (c advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
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
	return c.Parent().Permitted(m)
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

func (advancedTalk) Parent() core.CommandStatic {
	return Advanced
}

func (advancedTalk) Children() core.CommandsStatic {
	return nil
}

func (advancedTalk) Init() error {
	return nil
}

func (c advancedTalk) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTalk) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	resp, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: resp,
	}
	return embed, nil, nil
}

func (c advancedTalk) text(m *core.Message) (string, error, error) {
	resp, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return resp, nil, nil
}

func (advancedTalk) core(m *core.Message) (string, error) {
	return Talk(m.RawArgs(0))
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

func (c advancedReply) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
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

func (advancedReplyShow) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyShow) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyShow) Init() error {
	return nil
}

func (c advancedReplyShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	on, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(on),
	}
	return embed, nil, nil
}

func (c advancedReplyShow) text(m *core.Message) (string, error, error) {
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

func (advancedReplyOn) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyOn) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyOn) Init() error {
	return nil
}

func (c advancedReplyOn) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyOn) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedReplyOn) text(m *core.Message) (string, error, error) {
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

func (advancedReplyOff) Parent() core.CommandStatic {
	return AdvancedReply
}

func (advancedReplyOff) Children() core.CommandsStatic {
	return nil
}

func (advancedReplyOff) Init() error {
	return nil
}

func (c advancedReplyOff) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReplyOff) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedReplyOff) text(m *core.Message) (string, error, error) {
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

func (advancedInterval) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
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

func (advancedIntervalShow) Parent() core.CommandStatic {
	return AdvancedInterval
}

func (advancedIntervalShow) Children() core.CommandsStatic {
	return nil
}

func (advancedIntervalShow) Init() error {
	return nil
}

func (c advancedIntervalShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedIntervalShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	interval, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval),
	}
	return embed, nil, nil
}

func (c advancedIntervalShow) text(m *core.Message) (string, error, error) {
	interval, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval), nil, nil
}

func (advancedIntervalShow) fmt(interval time.Duration) string {
	return "The interval is set to: " + interval.String()
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
	return []string{
		"set",
	}
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

func (advancedIntervalSet) Parent() core.CommandStatic {
	return AdvancedInterval
}

func (advancedIntervalSet) Children() core.CommandsStatic {
	return nil
}

func (advancedIntervalSet) Init() error {
	return nil
}

func (c advancedIntervalSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedIntervalSet) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	interval, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(interval, usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedIntervalSet) text(m *core.Message) (string, error, error) {
	interval, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(interval, usrErr), usrErr, nil
}

func (advancedIntervalSet) fmt(interval time.Duration, usrErr error) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Updated the interval to %s.", interval)
	case ErrIntervalTooShort:
		return fmt.Sprintf("The interval %s is too short, must be longer or equal to %s.", interval, core.MinGodInterval)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedIntervalSet) core(m *core.Message) (time.Duration, error, error) {
	seconds, err := strconv.ParseInt(m.Command.Args[0], 10, 64)
	if err != nil {
		return time.Second, ErrInvalidInterval, nil
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Second, nil, err
	}
	interval := time.Duration(seconds) * time.Second
	usrErr, err := ReplyIntervalSet(here, interval)
	return interval, usrErr, err
}
