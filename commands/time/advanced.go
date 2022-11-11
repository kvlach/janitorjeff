package time

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/commands/nick"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var (
	errPersonNotFound  = errors.New("was unable to find user")
	errInvalidRemindID = errors.New("invalid reminder ID")
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.Type {
	return core.Advanced
}

func (advanced) Frontends() int {
	return frontends.All
}

func (advanced) Names() []string {
	return []string{
		"time",
	}
}

func (advanced) Description() string {
	return "Time stuff and things."
}

func (advanced) UsageArgs() string {
	return "(now | convert | timestamp | timezone | remind)"
}

func (advanced) Parent() core.Commander {
	return nil
}

func (advanced) Children() core.Commanders {
	return core.Commanders{
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

	return core.Globals.DB.Init(dbSchema)
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////
//     //
// now //
//     //
/////////

var AdvancedNow = advancedNow{}

type advancedNow struct{}

func (c advancedNow) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedNow) Frontends() int {
	return c.Parent().Frontends()
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

func (advancedNow) Parent() core.Commander {
	return Advanced
}

func (advancedNow) Children() core.Commanders {
	return nil
}

func (advancedNow) Init() error {
	return nil
}

func (c advancedNow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedNow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	now, cmdTzSet, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	cmdTzSet = discord.PlaceInBackticks(cmdTzSet)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, m, now, cmdTzSet),
	}

	return embed, usrErr, nil
}

func (c advancedNow) text(m *core.Message) (string, error, error) {
	now, cmdTzSet, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	cmdTzSet = fmt.Sprintf("'%s'", cmdTzSet)
	return c.err(usrErr, m, now, cmdTzSet), usrErr, nil
}

func (advancedNow) err(usrErr error, m *core.Message, now time.Time, cmdTzSet string) string {
	switch usrErr {
	case nil:
		return now.Format(time.RFC1123)
	case errTimezoneNotSet:
		return fmt.Sprintf("User %s has not set their timezone, to set a timezone use the %s command.", m.User.Mention, cmdTzSet)
	case errPersonNotFound:
		return fmt.Sprintf("Was unable to find the user %s", m.Command.Args[0])
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedNow) core(m *core.Message) (time.Time, string, error, error) {
	cmdTzSet := core.Format(AdvancedTimezoneSet, m.Command.Prefix)

	var person int64
	var err error
	if len(m.Command.Args) == 0 {
		person, err = m.Author()
	} else {
		person, err = nick.ParsePersonHere(m, m.Command.Args[0])
	}

	if err != nil {
		return time.Time{}, cmdTzSet, errPersonNotFound, nil
	}

	here, err := m.HereLogical()
	if err != nil {
		return time.Time{}, cmdTzSet, nil, err
	}

	now, usrErr, err := runNow(person, here)
	return now, cmdTzSet, usrErr, err
}

/////////////
//         //
// convert //
//         //
/////////////

var AdvancedConvert = advancedConvert{}

type advancedConvert struct{}

func (c advancedConvert) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedConvert) Frontends() int {
	return c.Parent().Frontends()
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

func (advancedConvert) Parent() core.Commander {
	return Advanced
}

func (advancedConvert) Children() core.Commanders {
	return nil
}

func (advancedConvert) Init() error {
	return nil
}

func (c advancedConvert) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedConvert) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, t),
	}

	return embed, usrErr, nil
}

func (c advancedConvert) text(m *core.Message) (string, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr, t), usrErr, nil
}

func (advancedConvert) err(usrErr error, t string) string {
	switch usrErr {
	case nil:
		return t
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedConvert) core(m *core.Message) (string, error, error) {
	target := m.Command.Args[0]
	tz := m.Command.Args[1]
	return runConvert(target, tz)
}

///////////////
//           //
// timestamp //
//           //
///////////////

var AdvancedTimestamp = advancedTimestamp{}

type advancedTimestamp struct{}

func (c advancedTimestamp) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedTimestamp) Frontends() int {
	return c.Parent().Frontends()
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

func (advancedTimestamp) Parent() core.Commander {
	return Advanced
}

func (advancedTimestamp) Children() core.Commanders {
	return nil
}

func (advancedTimestamp) Init() error {
	return nil
}

func (c advancedTimestamp) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimestamp) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, t),
	}

	if usrErr != nil {
		return embed, usrErr, nil
	}

	embed.Footer = &dg.MessageEmbedFooter{
		Text: t.Format(time.RFC1123),
	}

	return embed, nil, nil
}

func (c advancedTimestamp) text(m *core.Message) (string, error, error) {
	t, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr, t), usrErr, nil
}

func (advancedTimestamp) err(usrErr error, t time.Time) string {
	switch usrErr {
	case nil:
		return fmt.Sprint(t.Unix())
	case errInvalidTime:
		return "I can't understand what date that is."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedTimestamp) core(m *core.Message) (time.Time, error, error) {
	author, err := m.Author()
	if err != nil {
		return time.Time{}, nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return time.Time{}, nil, err
	}

	when := m.RawArgs(0)

	return runTimestamp(when, author, here)
}

//////////////
//          //
// timezone //
//          //
//////////////

var AdvancedTimezone = advancedTimezone{}

type advancedTimezone struct{}

func (c advancedTimezone) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedTimezone) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedTimezone) Names() []string {
	return []string{
		"timezone",
		"zone",
	}
}

func (advancedTimezone) Description() string {
	return "Show, set or delete your timezone."
}

func (advancedTimezone) UsageArgs() string {
	return "(show | set | delete)"
}

func (advancedTimezone) Parent() core.Commander {
	return Advanced
}

func (advancedTimezone) Children() core.Commanders {
	return core.Commanders{
		AdvancedTimezoneShow,
		AdvancedTimezoneSet,
		AdvancedTimezoneDelete,
	}
}

func (advancedTimezone) Init() error {
	return nil
}

func (advancedTimezone) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

///////////////////
//               //
// timezone show //
//               //
///////////////////

var AdvancedTimezoneShow = advancedTimezoneShow{}

type advancedTimezoneShow struct{}

func (c advancedTimezoneShow) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedTimezoneShow) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedTimezoneShow) Names() []string {
	return core.Show
}

func (advancedTimezoneShow) Description() string {
	return "Show the timezone that you set."
}

func (advancedTimezoneShow) UsageArgs() string {
	return ""
}

func (advancedTimezoneShow) Parent() core.Commander {
	return AdvancedTimezone
}

func (advancedTimezoneShow) Children() core.Commanders {
	return nil
}

func (advancedTimezoneShow) Init() error {
	return nil
}

func (c advancedTimezoneShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimezoneShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, tz),
	}

	return embed, usrErr, nil
}

func (c advancedTimezoneShow) text(m *core.Message) (string, error, error) {
	tz, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return c.err(usrErr, tz), usrErr, nil
}

func (advancedTimezoneShow) err(usrErr error, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your timezone is: %s", tz)
	case errTimezoneNotSet:
		return "Your timezone was not found."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedTimezoneShow) core(m *core.Message) (string, error, error) {
	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	return runTimezoneShow(author, here)
}

//////////////////
//              //
// timezone set //
//              //
//////////////////

var AdvancedTimezoneSet = advancedTimezoneSet{}

type advancedTimezoneSet struct{}

func (c advancedTimezoneSet) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedTimezoneSet) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedTimezoneSet) Names() []string {
	return []string{
		"set",
	}
}

func (advancedTimezoneSet) Description() string {
	return "Set your timezone."
}

func (advancedTimezoneSet) UsageArgs() string {
	return "<timezone>"
}

func (advancedTimezoneSet) Parent() core.Commander {
	return AdvancedTimezone
}

func (advancedTimezoneSet) Children() core.Commanders {
	return nil
}

func (advancedTimezoneSet) Init() error {
	return nil
}

func (c advancedTimezoneSet) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimezoneSet) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, m, tz),
	}

	return embed, usrErr, nil
}

func (c advancedTimezoneSet) text(m *core.Message) (string, error, error) {
	tz, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return c.err(usrErr, m, tz), usrErr, nil
}

func (advancedTimezoneSet) err(usrErr error, m *core.Message, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added %s with timezone %s", m.User.Mention, tz)
	case errTimezone:
		return fmt.Sprintf("%s is not a valid timezone.", tz)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedTimezoneSet) core(m *core.Message) (string, error, error) {
	tz := m.Command.Args[0]

	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	return runTimezoneSet(tz, author, here)
}

/////////////////////
//                 //
// timezone delete //
//                 //
/////////////////////

var AdvancedTimezoneDelete = advancedTimezoneDelete{}

type advancedTimezoneDelete struct{}

func (c advancedTimezoneDelete) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedTimezoneDelete) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedTimezoneDelete) Names() []string {
	return core.Delete
}

func (advancedTimezoneDelete) Description() string {
	return "Delete the timezone that you set."
}

func (advancedTimezoneDelete) UsageArgs() string {
	return ""
}

func (advancedTimezoneDelete) Parent() core.Commander {
	return AdvancedTimezone
}

func (advancedTimezoneDelete) Children() core.Commanders {
	return nil
}

func (advancedTimezoneDelete) Init() error {
	return nil
}

func (c advancedTimezoneDelete) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedTimezoneDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, m),
	}

	return embed, usrErr, nil
}

func (c advancedTimezoneDelete) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr, m), usrErr, nil
}

func (advancedTimezoneDelete) err(usrErr error, m *core.Message) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Deleted timezone for user %s", m.User.Mention)
	case errTimezoneNotSet:
		return fmt.Sprintf("Can't delete, user %s hasn't set their timezone.", m.User.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedTimezoneDelete) core(m *core.Message) (error, error) {
	author, err := m.Author()
	if err != nil {
		return nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	return runTimezoneDelete(author, here)
}

////////////
//        //
// remind //
//        //
////////////

var AdvancedRemind = advancedRemind{}

type advancedRemind struct{}

func (c advancedRemind) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedRemind) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedRemind) Names() []string {
	return []string{
		"remind",
	}
}

func (advancedRemind) Description() string {
	return "Reminder related commands."
}

func (advancedRemind) UsageArgs() string {
	return "(add | delete | list)"
}

func (advancedRemind) Parent() core.Commander {
	return Advanced
}

func (advancedRemind) Children() core.Commanders {
	return core.Commanders{
		AdvancedRemindAdd,
		AdvancedRemindDelete,
		AdvancedRemindList,
	}
}

func (advancedRemind) Init() error {
	return nil
}

func (advancedRemind) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////////
//            //
// remind add //
//            //
////////////////

var AdvancedRemindAdd = advancedRemindAdd{}

type advancedRemindAdd struct{}

func (c advancedRemindAdd) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedRemindAdd) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedRemindAdd) Names() []string {
	return core.Add
}

func (advancedRemindAdd) Description() string {
	return "Create a reminder."
}

func (advancedRemindAdd) UsageArgs() string {
	return "(<person> to <what> in <when> | <person> in <when> to <what>)"
}

func (advancedRemindAdd) Parent() core.Commander {
	return AdvancedRemind
}

func (advancedRemindAdd) Children() core.Commanders {
	return nil
}

func (advancedRemindAdd) Init() error {
	return nil
}

func (c advancedRemindAdd) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	t, id, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	return fmt.Sprintf("%s (#%d)", t.Format(time.RFC1123), id), nil, nil
}

func (advancedRemindAdd) core(m *core.Message) (time.Time, int64, error, error) {
	rxPerson := `(?P<person>[^\s]+)`
	rxWhen := `(in|on)\s+(?P<when>.+)`
	rxWhat := `to\s+(?P<what>.+)`

	re := regexp.MustCompile(`^` + rxPerson + `\s+` + rxWhat + `\s+` + rxWhen + `$`)
	groupNames := re.SubexpNames()

	var when string
	var what string
	var who string

	for _, match := range re.FindAllStringSubmatch(m.RawArgs(0), -1) {
		for i, text := range match {
			group := groupNames[i]

			switch group {
			case "when":
				when = text
			case "what":
				what = text
			case "person":
				who = text
			}
		}
	}

	person, err := nick.ParsePersonHere(m, who)
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	hereExact, err := m.HereExact()
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	hereLogical, err := m.HereLogical()
	if err != nil {
		return time.Time{}, -1, nil, err
	}

	return runRemindAdd(when, what, m.ID, person, hereExact, hereLogical)
}

///////////////////
//               //
// remind delete //
//               //
///////////////////

var AdvancedRemindDelete = advancedRemindDelete{}

type advancedRemindDelete struct{}

func (c advancedRemindDelete) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedRemindDelete) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedRemindDelete) Names() []string {
	return core.Delete
}

func (advancedRemindDelete) Description() string {
	return "Delete a reminder."
}

func (advancedRemindDelete) UsageArgs() string {
	return "<id>"
}

func (advancedRemindDelete) Parent() core.Commander {
	return AdvancedRemind
}

func (advancedRemindDelete) Children() core.Commanders {
	return nil
}

func (advancedRemindDelete) Init() error {
	return nil
}

func (c advancedRemindDelete) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedRemindDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}

	return embed, usrErr, nil
}

func (c advancedRemindDelete) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedRemindDelete) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Deleted reminder."
	case errReminderNotFound:
		return "Reminder not found. Maybe you are not the one who created the reminder?"
	case errInvalidRemindID:
		return "The ID you provided is invalid, expected a number."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedRemindDelete) core(m *core.Message) (error, error) {
	id, err := strconv.ParseInt(m.Command.Args[0], 10, 64)
	if err != nil {
		return errInvalidRemindID, nil
	}

	author, err := m.Author()
	if err != nil {
		return nil, err
	}

	return runRemindDelete(id, author)
}

/////////////////
//             //
// remind list //
//             //
/////////////////

var AdvancedRemindList = advancedRemindList{}

type advancedRemindList struct{}

func (c advancedRemindList) Type() core.Type {
	return c.Parent().Type()
}

func (c advancedRemindList) Frontends() int {
	return c.Parent().Frontends()
}

func (advancedRemindList) Names() []string {
	return core.List
}

func (advancedRemindList) Description() string {
	return "List active reminders."
}

func (advancedRemindList) UsageArgs() string {
	return ""
}

func (advancedRemindList) Parent() core.Commander {
	return AdvancedRemind
}

func (advancedRemindList) Children() core.Commanders {
	return nil
}

func (advancedRemindList) Init() error {
	return nil
}

func (c advancedRemindList) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return nil, nil, nil
	}
}

func (c advancedRemindList) discord(m *core.Message) (string, error, error) {
	rs, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
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

func (advancedRemindList) core(m *core.Message) ([]reminder, error, error) {
	author, err := m.Author()
	if err != nil {
		return nil, nil, err
	}

	here, err := m.HereExact()
	if err != nil {
		return nil, nil, err
	}

	return runRemindList(author, here)
}
