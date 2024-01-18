package paintball

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.Message) bool {
	if m.Frontend.Type() != discord.Frontend.Type() {
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
	game = createGame()
	movies = readMovies()
	fakeMovies = readFakeMovies()
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

func (c normal) Run(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return nil, nil, fmt.Errorf("Discord only")
	}
}

func (c normal) discord(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return normalHelp, core.UrrMissingArgs, nil
	}
	return c.play(m)
}

func (c normal) play(m *core.Message) (*dg.MessageEmbed, core.Urr, error) {
	rounds, err := strconv.Atoi(m.Command.Args[0])
	if err != nil {
		return normalHelp, nil, nil
	}
	if rounds < 1 || rounds > 15 {
		return normalHelp, nil, nil
	}

	here, err := m.Here.ScopeExact()
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
	hix, err := m.Here.IDExact()
	if err != nil {
		return nil, nil, err
	}
	return c.playF(hix, here, rounds)
}

func (normal) playF(channel string, here int64, rounds int) (*dg.MessageEmbed, core.Urr, error) {
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
		var err error
		if answer == nil {
			name = ""
		} else {
			name, err = answer.Author.DisplayName()
			if err != nil {
				log.Error().Err(err).Msg("failed to get author display name")
				return nil, nil, err
			}
			game.Point(here, answer.Author)
		}

		resp := generateAnswer(r, name, poster, answers, r == rounds)
		write(channel, resp)
	}

	write(channel, generateScorecard(game.Scores(here)))
	game.Playing(here, false)
	return nil, nil, core.UrrSilence
}
