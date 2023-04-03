package id

import (
	"errors"
	"fmt"

	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var errIDNotFound = errors.New("Couldn't find ID for the specified user. Does this user exist?")

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(*core.Message) bool {
	return true
}

func (normal) Names() []string {
	return []string{
		"id",
	}
}

func (normal) Description() string {
	return "Mention a user in some way and find their ID."
}

func (normal) UsageArgs() string {
	return "<user>"
}

func (normal) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (normal) Examples() []string {
	return nil
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return nil
}

func (normal) Init() error {
	return nil
}

func (c normal) Run(m *core.Message) (any, error, error) {
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

func (c normal) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	id, err := c.core(m)
	resp, usrErr := c.err(err, id)
	embed := &dg.MessageEmbed{
		Description: resp,
	}
	return embed, usrErr, nil
}

func (c normal) text(m *core.Message) (string, error, error) {
	id, err := c.core(m)
	resp, usrErr := c.err(err, id)
	return resp, usrErr, nil
}

func (normal) err(err error, id string) (string, error) {
	var usrErr error
	if err != nil {
		usrErr = errIDNotFound
	}

	switch usrErr {
	case nil:
		return id, nil
	default:
		return fmt.Sprint(usrErr), usrErr
	}
}

func (normal) core(m *core.Message) (string, error) {
	s := m.Command.Args[0]

	// in case a nick is used
	if person, err := nick.ParsePersonHere(m, s); err == nil {
		return core.DB.ScopeID(person)
	}

	return m.Client.PersonID(s, m.Here.ID())
}
