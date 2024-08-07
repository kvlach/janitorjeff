package id

import (
	"fmt"

	"github.com/kvlach/janitorjeff/commands/nick"
	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var UrrIDNotFound = core.UrrNew("Couldn't find ID for the specified user. Does this user exist?")

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(*core.EventMessage) bool {
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

func (c normal) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c normal) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	id, err := c.core(m)
	resp, urr := c.fmt(err, id)
	embed := &dg.MessageEmbed{
		Description: resp,
	}
	return embed, urr, nil
}

func (c normal) text(m *core.EventMessage) (string, core.Urr, error) {
	id, err := c.core(m)
	resp, urr := c.fmt(err, id)
	return resp, urr, nil
}

func (normal) fmt(err error, id string) (string, error) {
	var urr error
	if err != nil {
		urr = UrrIDNotFound
	}

	switch urr {
	case nil:
		return id, nil
	default:
		return fmt.Sprint(urr), urr
	}
}

func (normal) core(m *core.EventMessage) (string, error) {
	s := m.Command.Args[0]

	// in case a nick is used
	if person, err := nick.ParsePersonHere(m, s); err == nil {
		return core.DB.ScopeID(person)
	}
	hix, err := m.Here.IDExact()
	if err != nil {
		return "", err
	}
	return m.Client.PersonID(s, hix)
}
