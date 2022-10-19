package time

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/commands/nick"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var (
	errTimestamp      = errors.New("invalid timestamp")
	errTimezone       = errors.New("invalid timezone")
	errUserExists     = errors.New("user already exists")
	errTimezoneNotSet = errors.New("user hasn't set their timezone")
	errNotAuthor      = errors.New("only the user themselves can add their timezone")
	errUserNotFound   = errors.New("was unable to find user")
)

func runNormal(m *core.Message) (any, error, error) {
	return runNormalNow(m)
}

func runNormalTimezone(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) == 0 {
		return runNormalTimezoneGet(m)
	}
	return runNormalTimezoneSet(m)
}

func runNormalTimezoneSet(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runNormalTimezoneSetDiscord(m)
	default:
		return runNormalTimezoneSetText(m)
	}
}

func runNormalTimezoneSetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := runNormalTimezoneSetCore(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: runNormalTimezoneSetErr(usrErr, m, tz),
	}

	return embed, usrErr, nil
}

func runNormalTimezoneSetText(m *core.Message) (string, error, error) {
	tz, usrErr, err := runNormalTimezoneSetCore(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return runNormalTimezoneSetErr(usrErr, m, tz), usrErr, nil
}

func runNormalTimezoneSetErr(usrErr error, m *core.Message, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added %s with timezone %s", m.Author.Mention, tz)
	case errUserExists:
		// TODO: Suggest using '!time tz get' to view the set timezone
		return fmt.Sprintf("User %s already set their timezone.", m.Author.Mention)
	case errTimezone:
		return fmt.Sprintf("%s is not a valid timezone.", tz)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalTimezoneSetCore(m *core.Message) (string, error, error) {
	tz := m.Command.Runtime.Args[0]

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return tz, errTimezone, nil
	}

	tz = loc.String()

	user, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.ScopeHere()
	if err != nil {
		return "", nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return "", nil, err
	}
	if exists {
		return tz, errUserExists, nil
	}

	return tz, nil, dbUserAdd(user, place, tz)
}

func runNormalTimezoneDelete(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runNormalTimezoneDeleteDiscord(m)
	default:
		return runNormalTimezoneDeleteText(m)
	}
}

func runNormalTimezoneDeleteDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := runNormalTimezoneDeleteCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: runNormalTimezoneDeleteErr(usrErr, m),
	}

	return embed, usrErr, nil
}

func runNormalTimezoneDeleteText(m *core.Message) (string, error, error) {
	usrErr, err := runNormalTimezoneDeleteCore(m)
	if err != nil {
		return "", nil, err
	}
	return runNormalTimezoneDeleteErr(usrErr, m), usrErr, nil
}

func runNormalTimezoneDeleteErr(usrErr error, m *core.Message) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Deleted timezone for user %s", m.Author.Mention)
	case errTimezoneNotSet:
		return fmt.Sprintf("Can't delete, user %s hasn't set their timezone.", m.Author.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalTimezoneDeleteCore(m *core.Message) (error, error) {
	user, err := m.ScopeAuthor()
	if err != nil {
		return nil, err
	}

	place, err := m.ScopeHere()
	if err != nil {
		return nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errTimezoneNotSet, nil
	}

	return nil, dbUserDelete(user, place)
}

func runNormalTimezoneGet(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runNormalTimezoneGetDiscord(m)
	default:
		return runNormalTimezoneGetText(m)
	}
}

func runNormalTimezoneGetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	tz, usrErr, err := runNormalTimezoneGetCore(m)
	if err != nil {
		return nil, nil, err
	}

	tz = discord.PlaceInBackticks(tz)

	embed := &dg.MessageEmbed{
		Description: runNormalTimezoneGetErr(usrErr, m, tz),
	}

	return embed, usrErr, nil
}

func runNormalTimezoneGetText(m *core.Message) (string, error, error) {
	tz, usrErr, err := runNormalTimezoneGetCore(m)
	if err != nil {
		return "", nil, err
	}
	tz = fmt.Sprintf("'%s'", tz)
	return runNormalTimezoneGetErr(usrErr, m, tz), usrErr, nil
}

func runNormalTimezoneGetErr(usrErr error, m *core.Message, tz string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("User's %s timezone is: %s", m.Author.Mention, tz)
	case errTimezoneNotSet:
		return fmt.Sprintf("User's %s timezone was not found.", m.Author.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalTimezoneGetCore(m *core.Message) (string, error, error) {
	user, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.ScopeHere()
	if err != nil {
		return "", nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errTimezoneNotSet, nil
	}

	tz, err := dbUserTimezone(user, place)
	if err != nil {
		return "", nil, err
	}

	return tz, nil, nil
}

func runNormalConvert(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runNormalConvertDiscord(m)
	default:
		return runNormalConvertText(m)
	}
}

func runNormalConvertDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	t, usrErr, err := runNormalConvertCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: runNormalConvertErr(usrErr, t),
	}

	return embed, usrErr, nil
}

func runNormalConvertText(m *core.Message) (string, error, error) {
	t, usrErr, err := runNormalConvertCore(m)
	if err != nil {
		return "", nil, err
	}
	return runNormalConvertErr(usrErr, t), usrErr, nil
}

func runNormalConvertErr(usrErr error, t string) string {
	switch usrErr {
	case nil:
		return t
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalConvertCore(m *core.Message) (string, error, error) {
	target := m.Command.Runtime.Args[0]
	tz := m.Command.Runtime.Args[1]

	var t time.Time
	if target == "now" {
		t = time.Now().UTC()
	} else {
		timestamp, err := strconv.ParseInt(target, 10, 64)
		if err != nil {
			return "", errTimestamp, nil
		}
		t = time.Unix(timestamp, 0).UTC()
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", errTimezone, nil
	}
	return t.In(loc).Format(time.UnixDate), nil, nil
}

func runNormalNow(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runNormalNowDiscord(m)
	default:
		return runNormalNowText(m)
	}
}

func runNormalNowDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	now, cmdTzSet, usrErr, err := runNormalNowCore(m)
	if err != nil {
		return nil, nil, err
	}

	cmdTzSet = discord.PlaceInBackticks(cmdTzSet)

	embed := &dg.MessageEmbed{
		Description: runNormalNowErr(usrErr, m, now, cmdTzSet),
		Footer: &dg.MessageEmbedFooter{
			Text: fmt.Sprintf("Unix timestamp: %d", now.Unix()),
		},
	}

	return embed, usrErr, nil
}

func runNormalNowText(m *core.Message) (string, error, error) {
	now, cmdTzSet, usrErr, err := runNormalNowCore(m)
	if err != nil {
		return "", nil, err
	}
	cmdTzSet = fmt.Sprintf("'%s'", cmdTzSet)
	return runNormalNowErr(usrErr, m, now, cmdTzSet), usrErr, nil
}

func runNormalNowErr(usrErr error, m *core.Message, now time.Time, cmdTzSet string) string {
	switch usrErr {
	case nil:
		return now.Format(time.RFC1123)
	case errTimezoneNotSet:
		return fmt.Sprintf("User %s has not set their timezone, to set a timezone use the %s command.", m.Author.Mention, cmdTzSet)
	case errUserNotFound:
		return fmt.Sprintf("Was unable to find the user %s", m.Command.Runtime.Args[0])
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalNowCore(m *core.Message) (time.Time, string, error, error) {
	now := time.Now().UTC()
	cmdTzSet := cmdNormalTimezone.Format(m.Command.Runtime.Prefix)

	var user int64
	var err error
	if len(m.Command.Runtime.Args) == 0 {
		user, err = m.ScopeAuthor()
	} else {
		user, err = nick.ParseUserHere(m, m.Command.Runtime.Args[0])
	}

	if err != nil {
		return now, cmdTzSet, errUserNotFound, nil
	}

	place, err := m.ScopeHere()
	if err != nil {
		return now, cmdTzSet, nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return now, cmdTzSet, nil, err
	}

	if !exists {
		return now, cmdTzSet, errTimezoneNotSet, nil
	}

	tz, err := dbUserTimezone(user, place)
	if err != nil {
		return now, cmdTzSet, nil, err
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return now, cmdTzSet, nil, err
	}

	return now.In(loc), cmdTzSet, nil, nil
}
