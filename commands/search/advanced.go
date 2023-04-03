package search

import (
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.Message) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"search",
	}
}

func (advanced) Description() string {
	return "Search through the commands."
}

func (advanced) UsageArgs() string {
	return "<query...>"
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
	return nil
}

func (advanced) Init() error {
	return nil
}

func (c advanced) Run(m *core.Message) (any, error, error) {
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

func (c advanced) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	matches := c.core(m)

	var desc strings.Builder
	for i, match := range matches {
		cmd := match.command

		desc.WriteString("`")
		desc.WriteString(core.Format(cmd, m.Command.Prefix))
		desc.WriteString("`\n")
		desc.WriteString(cmd.Description())

		if i != len(matches)-1 {
			desc.WriteString("\n\n")
		}
	}

	embed := &dg.MessageEmbed{
		Description: desc.String(),
	}

	return embed, nil, nil
}

func (c advanced) text(m *core.Message) (string, error, error) {
	matches := c.core(m)

	var b strings.Builder
	for i, match := range matches {
		b.WriteString(core.Format(match.command, m.Command.Prefix))

		if i != len(matches)-1 {
			b.WriteString(" █ ")
		}
	}

	return b.String(), nil, nil
}

func (c advanced) core(m *core.Message) []Match {
	return Search(m.RawArgs(0), c.Type())
}
