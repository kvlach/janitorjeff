package paintball

import (
	"fmt"
	"strconv"
	"time"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.Message) bool {
	if m.Frontend != frontends.Discord {
		return false
	}
	return true
}

func (normal) Names() []string {
	return []string{
		"pb",
	}
}

func (normal) Description() string {
	return "Paintball game."
}

func (normal) UsageArgs() string {
	return "<rounds>"
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

var normalHelp = &dg.MessageEmbed{
	Title: "⚔ 🔫 **Welcome to Paintball!** 🔫 ⚔",
	Description: "❓ **__How to Play__** ❓\n" +
		"Paintball is a game of speed and knowledge.\n" +
		"A question will appear on screen.\n" +
		"Be the first to answer the question correctly!\n" +
		"The person who has won the most rounds wins!\n" +
		"Play with others to make the game more challenging!\n" +
		"\n" +
		"⌨ **__Start a Game__** ⌨\n" +
		"🥊 `!pb <rounds>` - To play a game of **UP TO 15** rounds.",
	Color: embedColor,
}

func (c normal) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return nil, nil, fmt.Errorf("Discord only")
	}
}

func (c normal) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	if len(m.Command.Args) < 1 {
		return normalHelp, core.ErrMissingArgs, nil
	}
	return c.play(m)
}

func (c normal) play(m *core.Message) (*dg.MessageEmbed, error, error) {
	rounds, err := strconv.Atoi(m.Command.Args[0])
	if err != nil {
		return normalHelp, nil, nil
	}
	if rounds < 1 || rounds > 15 {
		return normalHelp, nil, nil
	}

	here, err := m.HereExact()
	if err != nil {
		return nil, nil, err
	}

	if game.Active(here) == true {
		resp := &dg.MessageEmbed{
			Description: "A game is already active.",
			Color:       embedColor,
		}
		return resp, nil, nil
	}

	game.Playing(here, true)
	return c.playF(m.Channel.ID(), here, rounds)
}

func (normal) playF(channel string, here int64, rounds int) (*dg.MessageEmbed, error, error) {
	write(channel, &dg.MessageEmbed{
		Title:       "🔥 **Free-For-All** 🔥",
		Description: "Game starting in a few seconds!",
	})

	for r := 1; r <= rounds; r++ {
		// give some time for information to be processed by the players
		time.Sleep(interval)

		question, answers, poster := generateQuestion(r)
		write(channel, question)

		answer := awaitAnswer(here, answers)
		var name string
		if answer == nil {
			name = ""
		} else {
			name = answer.Author.DisplayName()
			game.Point(here, answer.Author)
		}

		resp := generateAnswer(r, name, poster, answers, r == rounds)
		write(channel, resp)
	}

	write(channel, generateScorecard(game.Scores(here)))
	game.Playing(here, false)
	return nil, nil, core.ErrSilence
}
