package rps

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var UrrUnexpectedArgument = core.UrrNew("got an unexpected argument")

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
		"rps",
	}
}

func (normal) Description() string {
	return "Rock paper scissors."
}

func (normal) UsageArgs() string {
	return "(r[ock] | p[aper] | s[cissors])"
}

func (normal) Category() core.CommandCategory {
	return core.CommandCategoryGames
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
	result, computer, urr := c.core(m)
	if urr != nil {
		return &dg.MessageEmbed{Description: c.fmt(urr)}, urr, nil
	}

	var title string
	switch result {
	case draw:
		title = "Draw."
	case win:
		title = "You win! 🥳"
	case loss:
		title = "You lost. 😦"
	}

	var desc string
	switch computer {
	case rock:
		desc = "I chose rock. 🪨"
	case paper:
		desc = "I chose paper. 🧻"
	case scissors:
		desc = "I chose scissors. ✂"
	}

	embed := &dg.MessageEmbed{
		Title:       title,
		Description: desc,
	}

	return embed, nil, nil
}

func (c normal) text(m *core.EventMessage) (string, core.Urr, error) {
	result, computer, urr := c.core(m)
	if urr != nil {
		return c.fmt(urr), urr, nil
	}

	var title string
	switch result {
	case draw:
		title = "Draw."
	case win:
		title = "You win!"
	case loss:
		title = "You lost."
	}

	var desc string
	switch computer {
	case rock:
		desc = "I chose rock."
	case paper:
		desc = "I chose paper."
	case scissors:
		desc = "I chose scissors."
	}

	return fmt.Sprintf("%s %s", title, desc), nil, nil
}

func (normal) fmt(urr core.Urr) string {
	switch urr {
	case UrrUnexpectedArgument:
		return "Please choose on of the following: rock, paper or scissors."
	default:
		return fmt.Sprint(urr)
	}
}

func (normal) core(m *core.EventMessage) (int, int, core.Urr) {
	var player int
	switch m.Command.Args[0] {
	case "r", "rock", "🪨":
		player = rock
	case "p", "paper", "🧻", "📰", "🗞 ":
		player = paper
	case "s", "scissors", "✂":
		player = scissors
	default:
		return -1, -1, UrrUnexpectedArgument
	}
	result, computer := run(player)
	return result, computer, nil
}
