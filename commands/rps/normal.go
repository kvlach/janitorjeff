package rps

import (
	"errors"
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"rps",
	},
	Description: "Rock paper scissors.",
	UsageArgs:   "(r[ock] | p[aper] | s[cissors])",
	Run:         normalRun,
}

var (
	errUnexpectedArgument = errors.New("got an unexpected argument")
)

func normalRun(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return normalRunDiscord(m)
	default:
		return normalRunText(m)
	}
}

func normalRunDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	result, computer, usrErr := normalRunCore(m)
	if usrErr != nil {
		return &dg.MessageEmbed{Description: normalRunErr(usrErr)}, usrErr, nil
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

func normalRunText(m *core.Message) (string, error, error) {
	result, computer, usrErr := normalRunCore(m)
	if usrErr != nil {
		return normalRunErr(usrErr), usrErr, nil
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

func normalRunErr(usrErr error) string {
	switch usrErr {
	case errUnexpectedArgument:
		return "Please choose on of the following: rock, paper or scissors."
	default:
		return fmt.Sprint(usrErr)
	}
}

func normalRunCore(m *core.Message) (int, int, error) {
	var player int
	switch m.Command.Runtime.Args[0] {
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
