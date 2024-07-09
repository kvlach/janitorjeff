package time

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kvlach/janitorjeff/commands/nick"
	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	UrrPersonNotFound  = core.UrrNew("was unable to find user")
	UrrInvalidRemindID = core.UrrNew("invalid reminder ID")
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.EventMessage) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"time",
	}
}

func (advanced) Description() string {
	return "Time stuff and things."
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
		AdvancedNow,
		AdvancedConvert,
		AdvancedTimestamp,
		AdvancedTimezone,
		AdvancedRemind,
	}
}

func (advanced) Init() error {
	go func() {
		for {
			runUpcoming()
			time.Sleep(2 * time.Minute)
		}
	}()

	return nil
}

func (advanced) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

/////////
//     //
// now //
//     //
/////////

var AdvancedNow = advancedNow{}

type advancedNow struct{}

func (c advancedNow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedNow) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedNow) Names() []string {
	return []string{
		"now",
	}
}

func (advancedNow) Description() string {
	return "View yours or someone else's time."
}

func (advancedNow) UsageArgs() string {
	return "[person]"
}

func (c advancedNow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedNow) Examples() []string {
	return nil
}

func (advancedNow) Parent() core.CommandStatic {
	return Advanced
}

func (advancedNow) Children() core.CommandsStatic {
	return nil
}

func (advancedNow) Init() error {
	return nil
}

func (c advancedNow) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedNow) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	now, cmdTzSet, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	cmdTzSet = discord.PlaceInBackticks(cmdTzSet)

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, m, now, cmdTzSet),
	}

	return embed, urr, nil
}

func (c advancedNow) text(m *core.EventMessage) (string, core.Urr, error) {
	now, cmdTzSet, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	cmdTzSet = fmt.Sprintf("'%s'", cmdTzSet)
	return c.fmt(urr, m, now, cmdTzSet), urr, nil
}

func (advancedNow) fmt(urr core.Urr, m *core.EventMessage, now time.Time, cmdTzSet string) string {
	switch urr {
	case nil:
		return now.Format(time.RFC1123)
	case UrrTimezoneNotSet:
		mention, err := m.Author.Mention()
		if err != nil {
			log.Error().Err(err).Msg("failed to get author mention")
			return ""
		}
		return fmt.Sprintf("User %s has not set their timezone, to set a timezone use the %s command.", mention, cmdTzSet)
	case UrrPersonNotFound:
		return fmt.Sprintf("Was unable to find the user %s", m.Command.Args[0])
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedNow) core(m *core.EventMessage) (time.Time, string, core.Urr, error) {
	cmdTzSet := core.Format(AdvancedTimezoneSet, m.Command.Prefix)

	var person int64
	var err error
	if len(m.Command.Args) == 0 {
		person, err = m.Author.Scope()
	} else {
		person, err = nick.ParsePersonHere(m, m.Command.Args[0])
	}

	if err != nil {
		return time.Time{}, cmdTzSet, UrrPersonNotFound, nil
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Time{}, cmdTzSet, nil, err
	}

	now, urr, err := Now(person, here)
	return now, cmdTzSet, urr, err
}

/////////////
//         //
// convert //
//         //
/////////////

var AdvancedConvert = advancedConvert{}

type advancedConvert struct{}

func (c advancedConvert) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedConvert) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedConvert) Names() []string {
	return []string{
		"convert",
	}
}

func (advancedConvert) Description() string {
	return "Convert a timestamp to the specified timezone."
}

func (advancedConvert) UsageArgs() string {
	return "<timestamp> <timezone>"
}

func (c advancedConvert) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedConvert) Examples() []string {
	return nil
}

func (advancedConvert) Parent() core.CommandStatic {
	return Advanced
}

func (advancedConvert) Children() core.CommandsStatic {
	return nil
}

func (advancedConvert) Init() error {
	return nil
}

func (c advancedConvert) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedConvert) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	t, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, t),
	}

	return embed, urr, nil
}

func (c advancedConvert) text(m *core.EventMessage) (string, core.Urr, error) {
	t, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr, t), urr, nil
}

func (advancedConvert) fmt(urr error, t string) string {
	switch urr {
	case nil:
		return t
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedConvert) core(m *core.EventMessage) (string, core.Urr, error) {
	target := m.Command.Args[0]
	tz := m.Command.Args[1]
	return Convert(target, tz)
}

///////////////
//           //
// timestamp //
//           //
///////////////

var AdvancedTimestamp = advancedTimestamp{}

type advancedTimestamp struct{}

func (c advancedTimestamp) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTimestamp) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedTimestamp) Names() []string {
	return []string{
		"timestamp",
	}
}

func (advancedTimestamp) Description() string {
	return "Get the given datetime's timestamp."
}

func (advancedTimestamp) UsageArgs() string {
	return "<when...>"
}

func (c advancedTimestamp) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTimestamp) Examples() []string {
	return nil
}

func (advancedTimestamp) Parent() core.CommandStatic {
	return Advanced
}

func (advancedTimestamp) Children() core.CommandsStatic {
	return nil
}

func (advancedTimestamp) Init() error {
	return nil
}

func (c advancedTimestamp) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedTimestamp) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	t, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, t),
	}

	if urr != nil {
		return embed, urr, nil
	}

	embed.Footer = &dg.MessageEmbedFooter{
		Text: t.Format(time.RFC1123),
	}

	return embed, nil, nil
}

func (c advancedTimestamp) text(m *core.EventMessage) (string, core.Urr, error) {
	t, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr, t), urr, nil
}

func (advancedTimestamp) fmt(urr core.Urr, t time.Time) string {
	switch urr {
	case nil:
		return fmt.Sprint(t.Unix())
	case UrrInvalidTime:
		return "I can't understand what date that is."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedTimestamp) core(m *core.EventMessage) (time.Time, core.Urr, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return time.Time{}, nil, err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Time{}, nil, err
	}

	when := m.RawArgs(0)

	return Timestamp(when, author, here)
}

//////////////
//          //
// timezone //
//          //
//////////////

var AdvancedTimezone = advancedTimezone{}

type advancedTimezone struct{}

func (c advancedTimezone) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTimezone) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedTimezone) Names() []string {
	return []string{
		"timezone",
		"zone",
		"tz",
	}
}

func (advancedTimezone) Description() string {
	return "Show, set or delete your timezone."
}

func (c advancedTimezone) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedTimezone) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTimezone) Examples() []string {
	return nil
}

func (advancedTimezone) Parent() core.CommandStatic {
	return Advanced
}

func (advancedTimezone) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedTimezoneShow,
		AdvancedTimezoneSet,
		AdvancedTimezoneDelete,
	}
}

func (advancedTimezone) Init() error {
	return nil
}

func (advancedTimezone) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

///////////////////
//               //
// timezone show //
//               //
///////////////////

var AdvancedTimezoneShow = advancedTimezoneShow{}

type advancedTimezoneShow struct{}

func (c advancedTimezoneShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTimezoneShow) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedTimezoneShow) Names() []string {
	return core.AliasesShow
}

func (advancedTimezoneShow) Description() string {
	return "Show the timezone that you set."
}

func (advancedTimezoneShow) UsageArgs() string {
	return ""
}

func (c advancedTimezoneShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTimezoneShow) Examples() []string {
	return nil
}

func (advancedTimezoneShow) Parent() core.CommandStatic {
	return AdvancedTimezone
}

func (advancedTimezoneShow) Children() core.CommandsStatic {
	return nil
}

func (advancedTimezoneShow) Init() error {
	return nil
}

func (c advancedTimezoneShow) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimezoneShow) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	tz, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: c.fmt(tz),
	}

	return embed, nil, nil
}

func (c advancedTimezoneShow) text(m *core.EventMessage) (string, core.Urr, error) {
	tz, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return c.fmt(tz), nil, nil
}

func (advancedTimezoneShow) fmt(tz string) string {
	return fmt.Sprintf("Your timezone is: %s", tz)
}

func (advancedTimezoneShow) core(m *core.EventMessage) (string, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return "", err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", err
	}

	return TimezoneShow(author, here)
}

//////////////////
//              //
// timezone set //
//              //
//////////////////

var AdvancedTimezoneSet = advancedTimezoneSet{}

type advancedTimezoneSet struct{}

func (c advancedTimezoneSet) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTimezoneSet) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedTimezoneSet) Names() []string {
	return core.AliasesSet
}

func (advancedTimezoneSet) Description() string {
	return "Set your timezone."
}

func (advancedTimezoneSet) UsageArgs() string {
	return "<timezone>"
}

func (c advancedTimezoneSet) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTimezoneSet) Examples() []string {
	return nil
}

func (advancedTimezoneSet) Parent() core.CommandStatic {
	return AdvancedTimezone
}

func (advancedTimezoneSet) Children() core.CommandsStatic {
	return nil
}

func (advancedTimezoneSet) Init() error {
	return nil
}

func (c advancedTimezoneSet) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedTimezoneSet) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	tz, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, m, tz),
	}

	return embed, urr, nil
}

func (c advancedTimezoneSet) text(m *core.EventMessage) (string, core.Urr, error) {
	tz, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return c.fmt(urr, m, tz), urr, nil
}

func (advancedTimezoneSet) fmt(urr core.Urr, m *core.EventMessage, tz string) string {
	switch urr {
	case nil:
		mention, err := m.Author.Mention()
		if err != nil {
			log.Error().Err(err).Msg("failed to get author mention")
			return ""
		}
		return fmt.Sprintf("Added %s with timezone %s", mention, tz)
	case UrrTimezone:
		return fmt.Sprintf("%s is not a valid timezone.", tz)
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedTimezoneSet) core(m *core.EventMessage) (string, core.Urr, error) {
	tz := m.Command.Args[0]

	author, err := m.Author.Scope()
	if err != nil {
		return "", nil, err
	}

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", nil, err
	}

	return TimezoneSet(tz, author, here)
}

/////////////////////
//                 //
// timezone delete //
//                 //
/////////////////////

var AdvancedTimezoneDelete = advancedTimezoneDelete{}

type advancedTimezoneDelete struct{}

func (c advancedTimezoneDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedTimezoneDelete) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedTimezoneDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedTimezoneDelete) Description() string {
	return "Delete the timezone that you set."
}

func (advancedTimezoneDelete) UsageArgs() string {
	return ""
}

func (c advancedTimezoneDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedTimezoneDelete) Examples() []string {
	return nil
}

func (advancedTimezoneDelete) Parent() core.CommandStatic {
	return AdvancedTimezone
}

func (advancedTimezoneDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedTimezoneDelete) Init() error {
	return nil
}

func (c advancedTimezoneDelete) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimezoneDelete) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(m),
	}
	return embed, nil, nil
}

func (c advancedTimezoneDelete) text(m *core.EventMessage) (string, core.Urr, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(m), nil, nil
}

func (advancedTimezoneDelete) fmt(m *core.EventMessage) string {
	mention, err := m.Author.Mention()
	if err != nil {
		log.Error().Err(err).Msg("failed to get author mention")
		return ""
	}
	return fmt.Sprintf("Deleted timezone for user %s", mention)
}

func (advancedTimezoneDelete) core(m *core.EventMessage) error {
	author, err := m.Author.Scope()
	if err != nil {
		return err
	}
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return TimezoneDelete(author, here)
}

////////////
//        //
// remind //
//        //
////////////

var AdvancedRemind = advancedRemind{}

type advancedRemind struct{}

func (c advancedRemind) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRemind) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedRemind) Names() []string {
	return []string{
		"remind",
	}
}

func (advancedRemind) Description() string {
	return "Reminder related commands."
}

func (c advancedRemind) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedRemind) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRemind) Examples() []string {
	return nil
}

func (advancedRemind) Parent() core.CommandStatic {
	return Advanced
}

func (advancedRemind) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedRemindAdd,
		AdvancedRemindDelete,
		AdvancedRemindList,
	}
}

func (advancedRemind) Init() error {
	return nil
}

func (advancedRemind) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

////////////////
//            //
// remind add //
//            //
////////////////

var AdvancedRemindAdd = advancedRemindAdd{}

type advancedRemindAdd struct{}

func (c advancedRemindAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRemindAdd) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedRemindAdd) Names() []string {
	return core.AliasesAdd
}

func (advancedRemindAdd) Description() string {
	return "Create a reminder."
}

func (advancedRemindAdd) UsageArgs() string {
	return "<what> (in|on) <when>"
}

func (c advancedRemindAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRemindAdd) Examples() []string {
	return nil
}

func (advancedRemindAdd) Parent() core.CommandStatic {
	return AdvancedRemind
}

func (advancedRemindAdd) Children() core.CommandsStatic {
	return nil
}

func (advancedRemindAdd) Init() error {
	return nil
}

func (c advancedRemindAdd) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	t, id, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}
	return fmt.Sprintf("%s (#%d)", t.Format(time.RFC1123), id), nil, nil
}

func (advancedRemindAdd) core(m *core.EventMessage) (time.Time, int64, core.Urr, error) {
	rxWhat := `(?P<what>.+)`
	rxWhen := `(in|on)\s+(?P<when>.+)`

	re := regexp.MustCompile(`^` + rxWhat + `\s+` + rxWhen + `$`)
	groupNames := re.SubexpNames()

	var when string
	var what string

	for _, match := range re.FindAllStringSubmatch(m.RawArgs(0), -1) {
		for i, text := range match {
			group := groupNames[i]

			switch group {
			case "when":
				when = text
			case "what":
				what = text
			}
		}
	}

	author, err := m.Author.Scope()
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	hereExact, err := m.Here.ScopeExact()
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	hereLogical, err := m.Here.ScopeLogical()
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	return RemindAdd(when, what, m.ID, author, hereExact, hereLogical)
}

///////////////////
//               //
// remind delete //
//               //
///////////////////

var AdvancedRemindDelete = advancedRemindDelete{}

type advancedRemindDelete struct{}

func (c advancedRemindDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRemindDelete) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedRemindDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedRemindDelete) Description() string {
	return "Delete a reminder."
}

func (advancedRemindDelete) UsageArgs() string {
	return "<id>"
}

func (c advancedRemindDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRemindDelete) Examples() []string {
	return nil
}

func (advancedRemindDelete) Parent() core.CommandStatic {
	return AdvancedRemind
}

func (advancedRemindDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedRemindDelete) Init() error {
	return nil
}

func (c advancedRemindDelete) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedRemindDelete) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}

	return embed, urr, nil
}

func (c advancedRemindDelete) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedRemindDelete) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Deleted reminder."
	case UrrReminderNotFound:
		return "Reminder not found. Maybe you are not the one who created the reminder?"
	case UrrInvalidRemindID:
		return "The ID you provided is invalid, expected a number."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedRemindDelete) core(m *core.EventMessage) (core.Urr, error) {
	id, err := strconv.ParseInt(m.Command.Args[0], 10, 64)
	if err != nil {
		return UrrInvalidRemindID, nil
	}

	author, err := m.Author.Scope()
	if err != nil {
		return nil, err
	}

	return RemindDelete(id, author)
}

/////////////////
//             //
// remind list //
//             //
/////////////////

var AdvancedRemindList = advancedRemindList{}

type advancedRemindList struct{}

func (c advancedRemindList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedRemindList) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedRemindList) Names() []string {
	return core.AliasesList
}

func (advancedRemindList) Description() string {
	return "List active reminders."
}

func (advancedRemindList) UsageArgs() string {
	return ""
}

func (c advancedRemindList) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedRemindList) Examples() []string {
	return nil
}

func (advancedRemindList) Parent() core.CommandStatic {
	return AdvancedRemind
}

func (advancedRemindList) Children() core.CommandsStatic {
	return nil
}

func (advancedRemindList) Init() error {
	return nil
}

func (c advancedRemindList) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return nil, nil, nil
	}
}

func (c advancedRemindList) discord(m *core.EventMessage) (string, core.Urr, error) {
	rs, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}

	var resp strings.Builder

	if len(rs) == 1 {
		fmt.Fprintf(&resp, "%d timer open.\n", len(rs))
	} else {
		fmt.Fprintf(&resp, "%d timers open.\n", len(rs))
	}

	now := time.Now()

	for _, r := range rs {
		remaining := r.When.Sub(now).Round(time.Second)
		fmt.Fprintf(&resp, "%d: %s (%s remaining)\n", r.ID, r.What, remaining)
	}

	return resp.String(), nil, nil
}

func (advancedRemindList) core(m *core.EventMessage) ([]reminder, core.Urr, error) {
	author, err := m.Author.Scope()
	if err != nil {
		return nil, nil, err
	}

	here, err := m.Here.ScopeExact()
	if err != nil {
		return nil, nil, err
	}

	return RemindList(author, here)
}
