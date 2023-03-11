package rps

import (
	"errors"
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var errUnexpectedArgument = errors.New("got an unexpected argument")

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
		"rps",
	}
}

func (normal) Description() string {
	return "Rock paper scissors."
}

func (normal) UsageArgs() string {
	return "(r[ock] | p[aper] | s[cissors])"
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
	result, computer, usrErr := c.core(m)
	if usrErr != nil {
		return &dg.MessageEmbed{Description: c.err(usrErr)}, usrErr, nil
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

func (c normal) text(m *core.Message) (string, error, error) {
	result, computer, usrErr := c.core(m)
	if usrErr != nil {
		return c.err(usrErr), usrErr, nil
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

func (normal) err(usrErr error) string {
	switch usrErr {
	case errUnexpectedArgument:
		return "Please choose on of the following: rock, paper or scissors."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (normal) core(m *core.Message) (int, int, error) {
	var player int
	switch m.Command.Args[0] {
	case "r", "rock", "🪨":
		player = rock
	case "p", "paper", "🧻", "📰", "🗞 ":
		player = paper
	case "s", "scissors", "✂":
		player = scissors
	default:
		return -1, -1, errUnexpectedArgument
	}
	result, computer := run(player)
	return result, computer, nil
}
