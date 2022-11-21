package test

import (
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.Message) bool {
	return true
}

func (admin) Names() []string {
	return []string{
		"test",
		"alias",
	}
}

func (admin) Description() string {
	return "Test command."
}

func (admin) UsageArgs() string {
	return ""
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return nil
}

func (admin) Init() error {
	return nil
}

func (c admin) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c admin) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	embed := &dg.MessageEmbed{
		Description: c.core(),
	}
	return embed, nil, nil
}

func (c admin) text(m *core.Message) (string, error, error) {
	return c.core(), nil, nil
}

func (admin) core() string {
	return "Test command!"
}
