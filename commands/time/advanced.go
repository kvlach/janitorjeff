package time

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/commands/nick"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var errPersonNotFound = errors.New("was unable to find user")

var Advanced = &core.CommandStatic{
	Names: []string{
		"time",
	},
	Description: "time stuff and things",
	UsageArgs:   "(now | convert | zone)",
	Run:         advancedRun,

	Children: core.Commands{
		{
			Names: []string{
				"now",
			},
			Description: "View yours or someone else's time.",
			UsageArgs:   "[person]",
			Run:         advancedRunNow,
		},
		{
			Names: []string{
				"convert",
			},
			Description: "Convert a timestamp to the specified timezone.",
			UsageArgs:   "<timestamp> <timezone>",
			Run:         advancedRunConvert,
		},
		{
			Names: []string{
				"timestamp",
			},
			Description: "Get the given datetime's timestamp.",
			UsageArgs:   "<when...>",
			Run:         advancedRunTimestamp,
		},
		{
			Names: []string{
				"zone",
			},
			Description: "View, set or delete your nickname.",
			UsageArgs:   "(view | set | delete)",
			Run:         advancedRunTimezone,

			Children: core.Commands{
				{
					Names: []string{
						"view",
						"get",
					},
					Description: "View the timezone that you set.",
					UsageArgs:   "",
					Run:         advancedRunTimezoneGet,
				},
				{
					Names: []string{
						"set",
					},
					Description: "Set your timezone.",
					UsageArgs:   "<timezone>",
					Run:         advancedRunTimezoneSet,
				},
				{
					Names: []string{
						"delete",
						"del",
						"remove",
						"rm",
					},
					Description: "Delete the timezone that you set.",
					UsageArgs:   "",
					Run:         advancedRunTimezoneDelete,
				},
			},
		},
		{
			Names: []string{
				"remind",
			},
			Description: "Reminder related commands",
			UsageArgs:   "(add)",
			Run:         advancedRunRemind,

			Children: core.Commands{
				{
					Names: []string{
						"add",
						"new",
						"create",
					},
					Description: "Create a reminder.",
					UsageArgs:   "(<person> to <what> in <when> | <person> in <when> to <what>)",
					Run:         advancedRunRemindAdd,
				},
			},
		},
	},
}

func advancedRun(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

/////////
//     //
// now //
//     //
/////////

func advancedRunNow(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunNowDiscord(m)
	default:
		return advancedRunNowText(m)
	}
}

func advancedRunNowDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	now, cmdTzSet, usrErr, err := advancedRunNowCore(m)
	if err != nil {
		return nil, nil, err
	}

	cmdTzSet = discord.PlaceInBackticks(cmdTzSet)

	embed := &dg.MessageEmbed{
		Description: advancedRunNowErr(usrErr, m, now, cmdTzSet),
	}

	return embed, usrErr, nil
}

func advancedRunNowText(m *core.Message) (string, error, error) {
	now, cmdTzSet, usrErr, err := advancedRunNowCore(m)
	if err != nil {
		return "", nil, err
	}
	cmdTzSet = fmt.Sprintf("'%s'", cmdTzSet)
	return advancedRunNowErr(usrErr, m, now, cmdTzSet), usrErr, nil
}

func advancedRunNowErr(usrErr error, m *core.Message, now time.Time, cmdTzSet string) string {
	switch usrErr {
	case nil:
		return now.Format(time.RFC1123)
	case errTimezoneNotSet:
		return fmt.Sprintf("User %s has not set their timezone, to set a timezone use the %s command.", m.Author.Mention, cmdTzSet)
	case errPersonNotFound:
		return fmt.Sprintf("Was unable to find the user %s", m.Command.Runtime.Args[0])
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunNowCore(m *core.Message) (time.Time, string, error, error) {
	cmdTzSet := cmdNormalTimezone.Format(m.Command.Runtime.Prefix)

	var person int64
	var err error
	if len(m.Command.Runtime.Args) == 0 {
		person, err = m.ScopeAuthor()
	} else {
		person, err = nick.ParsePersonHere(m, m.Command.Runtime.Args[0])
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

func advancedRunConvert(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return advancedRunConvertDiscord(m)
	default:
		return advancedRunConvertText(m)
	}
}

func advancedRunConvertDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := advancedRunConvertCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: advancedRunConvertErr(usrErr, t),
	}

	return embed, usrErr, nil
}

func advancedRunConvertText(m *core.Message) (string, error, error) {
	t, usrErr, err := advancedRunConvertCore(m)
	if err != nil {
		return "", nil, err
	}
	return advancedRunConvertErr(usrErr, t), usrErr, nil
}

func advancedRunConvertErr(usrErr error, t string) string {
	switch usrErr {
	case nil:
		return t
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunConvertCore(m *core.Message) (string, error, error) {
	target := m.Command.Runtime.Args[0]
	tz := m.Command.Runtime.Args[1]
	return runConvert(target, tz)
}

///////////////
//           //
// timestamp //
//           //
///////////////

func advancedRunTimestamp(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return advancedRunTimestampDiscord(m)
	default:
		return advancedRunTimestampText(m)
	}
}

func advancedRunTimestampDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := advancedRunTimestampCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: advancedRunTimestampErr(usrErr, t),
	}

	if usrErr != nil {
		return embed, usrErr, nil
	}

	embed.Footer = &dg.MessageEmbedFooter{
		Text: t.Format(time.RFC1123),
	}

	return embed, nil, nil
}

func advancedRunTimestampText(m *core.Message) (string, error, error) {
	t, usrErr, err := advancedRunTimestampCore(m)
	if err != nil {
		return "", nil, err
	}
	return advancedRunTimestampErr(usrErr, t), usrErr, nil
}

func advancedRunTimestampErr(usrErr error, t time.Time) string {
	switch usrErr {
	case nil:
		return fmt.Sprint(t.Unix())
	case errInvalidTime:
		return "I can't understand what date that is."
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunTimestampCore(m *core.Message) (time.Time, error, error) {
	author, err := m.ScopeAuthor()
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

func advancedRunTimezone(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

//////////////////
//              //
// timezone get //
//              //
//////////////////

func advancedRunTimezoneGet(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunTimezoneGetDiscord(m)
	default:
		return advancedRunTimezoneGetText(m)
	}
}

func advancedRunTimezoneGetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := advancedRunTimezoneGetCore(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: advancedRunTimezoneGetErr(usrErr, tz),
	}

	return embed, usrErr, nil
}

func advancedRunTimezoneGetText(m *core.Message) (string, error, error) {
	tz, usrErr, err := advancedRunTimezoneGetCore(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return advancedRunTimezoneGetErr(usrErr, tz), usrErr, nil
}

func advancedRunTimezoneGetErr(usrErr error, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your timezone is: %s", tz)
	case errTimezoneNotSet:
		return "Your timezone was not found."
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunTimezoneGetCore(m *core.Message) (string, error, error) {
	author, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	return runTimezoneGet(author, here)
}

//////////////////
//              //
// timezone set //
//              //
//////////////////

func advancedRunTimezoneSet(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return advancedRunTimezoneSetDiscord(m)
	default:
		return advancedRunTimezoneSetText(m)
	}
}

func advancedRunTimezoneSetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := advancedRunTimezoneSetCore(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: advancedRunTimezoneSetErr(usrErr, m, tz),
	}

	return embed, usrErr, nil
}

func advancedRunTimezoneSetText(m *core.Message) (string, error, error) {
	tz, usrErr, err := advancedRunTimezoneSetCore(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return advancedRunTimezoneSetErr(usrErr, m, tz), usrErr, nil
}

func advancedRunTimezoneSetErr(usrErr error, m *core.Message, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added %s with timezone %s", m.Author.Mention, tz)
	case errTimezone:
		return fmt.Sprintf("%s is not a valid timezone.", tz)
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunTimezoneSetCore(m *core.Message) (string, error, error) {
	tz := m.Command.Runtime.Args[0]

	author, err := m.ScopeAuthor()
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

func advancedRunTimezoneDelete(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunTimezoneDeleteDiscord(m)
	default:
		return advancedRunTimezoneDeleteText(m)
	}
}

func advancedRunTimezoneDeleteDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := advancedRunTimezoneDeleteCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: advancedRunTimezoneDeleteErr(usrErr, m),
	}

	return embed, usrErr, nil
}

func advancedRunTimezoneDeleteText(m *core.Message) (string, error, error) {
	usrErr, err := advancedRunTimezoneDeleteCore(m)
	if err != nil {
		return "", nil, err
	}
	return advancedRunTimezoneDeleteErr(usrErr, m), usrErr, nil
}

func advancedRunTimezoneDeleteErr(usrErr error, m *core.Message) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Deleted timezone for user %s", m.Author.Mention)
	case errTimezoneNotSet:
		return fmt.Sprintf("Can't delete, user %s hasn't set their timezone.", m.Author.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunTimezoneDeleteCore(m *core.Message) (error, error) {
	author, err := m.ScopeAuthor()
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

func advancedRunRemind(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

////////////////
//            //
// remind add //
//            //
////////////////

func advancedRunRemindAdd(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	t, id, usrErr, err := advancedRunRemindAddCore(m)
	if err != nil {
		return nil, nil, err
	}
	return fmt.Sprintf("%s (#%d)", t.Format(time.RFC1123), id), usrErr, nil
}

func advancedRunRemindAddCore(m *core.Message) (time.Time, int64, error, error) {
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

	return runRemindAdd(when, what, person, hereExact, hereLogical)
}
