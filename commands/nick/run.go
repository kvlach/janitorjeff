package nick

import (
	"errors"
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

var (
	errNotFound = errors.New("user nick not found")
	errExists   = errors.New("user nick has already been set")
)

func runNormal(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) == 0 {
		return runNormalGet(m)
	}
	return runNormalSet(m)
}

func runNormalGet(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runNormalGetDiscord(m)
	default:
		return runNormalGetText(m)
	}
}

func runNormalGetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := runNormalGetCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: runNormalGetErr(usrErr, m, nick),
	}

	return embed, usrErr, nil
}

func runNormalGetText(m *core.Message) (string, error, error) {
	nick, usrErr, err := runNormalGetCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return runNormalGetErr(usrErr, m, nick), usrErr, nil
}

func runNormalGetErr(usrErr error, m *core.Message, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("User's %s nickname is %s", m.Author.Mention, nick)
	case errNotFound:
		return fmt.Sprintf("User %s has not set their nickname.", m.Author.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalGetCore(m *core.Message) (string, error, error) {
	user, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.Scope()
	if err != nil {
		return "", nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errNotFound, nil
	}

	nick, err := dbUserNick(user, place)
	return nick, nil, err
}

func runNormalSet(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runNormalSetDiscord(m)
	default:
		return runNormalSetText(m)
	}
}

func runNormalSetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := runNormalSetCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: runNormalSetErr(usrErr, m, nick),
	}

	return embed, usrErr, nil
}

func runNormalSetText(m *core.Message) (string, error, error) {
	nick, usrErr, err := runNormalSetCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return runNormalSetErr(usrErr, m, nick), usrErr, nil
}

func runNormalSetErr(usrErr error, m *core.Message, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Set nickname %s for user %s", nick, m.Author.Mention)
	// case errExists:
	// 	return fmt.Sprintf("Can't set %s as nickname for user %s, one already exists.", nick, m.Author.Mention)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runNormalSetCore(m *core.Message) (string, error, error) {
	nick := m.Command.Runtime.Args[0]

	user, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.Scope()
	if err != nil {
		return "", nil, err
	}

	exists, err := dbUserExists(user, place)
	if err != nil {
		return "", nil, err
	}

	if exists {
		return nick, nil, dbUserUpdate(user, place, nick)
	}
	return nick, nil, dbUserAdd(user, place, nick)
}
