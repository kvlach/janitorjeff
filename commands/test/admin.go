package test

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.Type {
	return core.Admin
}

func (admin) Frontends() int {
	return frontends.All
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

func (admin) Parent() core.Commander {
	return nil
}

func (admin) Children() core.Commanders {
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
