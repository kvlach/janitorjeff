package nick

import (
	"errors"
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"nick",
		"nickname",
	},
	Description: "Set a nickname that can be used when calling commands.",
	UsageArgs:   "[nickname]",
	Run:         normalRun,
	Init:        init_,
}

var (
	errUserNotFound = errors.New("user nick not found")
	errNickExists   = errors.New("nick is used by a different user")
)

func normalRun(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) == 0 {
		return normalRunGet(m)
	}
	return normalRunSet(m)
}

func normalRunGet(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return normalRunGetDiscord(m)
	default:
		return normalRunGetText(m)
	}
}

func normalRunGetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := normalRunGetCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: normalRunGetErr(usrErr, nick),
	}

	return embed, usrErr, nil
}

func normalRunGetText(m *core.Message) (string, error, error) {
	nick, usrErr, err := normalRunGetCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return normalRunGetErr(usrErr, nick), usrErr, nil
}

func normalRunGetErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your nickname is: %s", nick)
	case errUserNotFound:
		return "You have not set a nickname."
	default:
		return fmt.Sprint(usrErr)
	}
}

func normalRunGetCore(m *core.Message) (string, error, error) {
	author, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.ScopeHere()
	if err != nil {
		return "", nil, err
	}

	return runGet(author, place)
}

func normalRunSet(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return normalRunSetDiscord(m)
	default:
		return normalRunSetText(m)
	}
}

func normalRunSetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := normalRunSetCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: normalRunSetErr(usrErr, nick),
	}

	return embed, usrErr, nil
}

func normalRunSetText(m *core.Message) (string, error, error) {
	nick, usrErr, err := normalRunSetCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return normalRunSetErr(usrErr, nick), usrErr, nil
}

func normalRunSetErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Nickname set to %s", nick)
	case errNickExists:
		return fmt.Sprintf("Nickname %s is already being used by another user.", nick)
	default:
		return fmt.Sprint(usrErr)
	}
}

func normalRunSetCore(m *core.Message) (string, error, error) {
	nick := m.Command.Runtime.Args[0]

	author, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.ScopeHere()
	if err != nil {
		return "", nil, err
	}

	usrErr, err := runSet(nick, author, place)
	return nick, usrErr, err
}

// Tries to find a user scope from the given string. First tries to find if it
// matches a nickname in the database and if it doesn't it tries various
// platform specific things, like for example checking if the given string is a
// user ID.
func ParseUser(m *core.Message, place int64, s string) (int64, error) {
	if user, err := dbGetUser(s, place); err == nil {
		return user, nil
	}

	placeID, err := core.Globals.DB.ScopeID(place)
	if err != nil {
		return -1, err
	}

	id, err := m.Client.PersonID(s, placeID)
	if err != nil {
		return -1, err
	}

	return m.Client.PersonScope(id)
}

// Same as ParseUser but uses the default place instead
func ParseUserHere(m *core.Message, s string) (int64, error) {
	place, err := m.ScopeHere()
	if err != nil {
		return -1, err
	}

	return ParseUser(m, place, s)
}
