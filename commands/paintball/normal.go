package paintball

import (
	"fmt"
	"strconv"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"pb",
	},
	Description: "Paintball game",
	UsageArgs:   "<play|help>",
	Run:         normalRun,

	Children: core.Commands{
		{
			Names: []string{
				"play",
			},
			Description: "Play the game.",
			UsageArgs:   "",
			Run:         normalRunPlay,
		},
	},
}

func normalRun(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return normalRunDiscord(m)
	default:
		return nil, nil, fmt.Errorf("Discord only")
	}
}

func normalRunDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	// desc := "An exclusive Filmtopia game...\n" +
	// 	"\n" +
	desc := "❓ **__How to Play__** ❓\n" +
		"Paintball is a game of speed and knowledge.\n" +
		"A question will appear on screen.\n" +
		"Be the first to answer the question correctly!\n" +
		"The person who has won the most rounds wins!\n" +
		"Play with others to make the game more challenging!\n" +
		"\n" +
		"⚔ **__Game Modes__** ⚔\n" +
		"🔥 **Free-For-All [f]** - Anyone can play at any time! You can even play single!\n" +
		// "👥 **Multiplayer [m]** - Play in group matches.\n" +
		// "🛡 **Teams [t]** - Play in teams to win! *Coming Soon*\n" +
		"\n" +
		"🎲 **__Question Types__** 🎲\n" +
		"🖼 **Poster** - Guess the movie name from the poster.\n" +
		"🧩 **Scramble** - Unscramble the movie name.\n" +
		"🔍 **Fake or Real** - Guess whether the movie is fake.\n" +
		"📆 **Year** - Guess the release year of the movie.\n" +
		"📣 **Director** - Guess the director of the movie.\n" +
		// "🤔 **True or False** - Guess whether the statement is true or false.\n" +
		"📖 **Plot** - Guess the movie from the plot.\n" +
		"\n" +
		"⌨ **__Commands__** ⌨\n" +
		"🥊 `!pb play` - To play a game.\n" +
		"❓ `!pb help` - For the help menu.\n"
		// "\n" +
		// "🎁 **__Prizes__** 🎁\n" +
		// "You win points when you win a game!\n" +
		// "You can also gain points even if you lose.\n" +
		// "**You can redeem your points for Discord Nitro**\n" +
		// "More information here: <https://shorturl.at/nsuyK>\n" +
		// "*You can only win points in Team matches.*\n" +
		// "\n" +
		// "⚠ **__Warning__** ⚠\n" +
		// "This bot is in beta.\n" +
		// "Suspicious use of this command will lead to a blacklist from the <#736577768652931202> channel.\n" +
		// "Harsher punishments will be imposed for continued conduct."

	help := &dg.MessageEmbed{
		Title:       "⚔ 🔫 **Welcome to Paintball!** 🔫 ⚔",
		Description: desc,
		Color:       embedColor,
	}

	return help, nil, nil
}

func normalRunPlay(m *core.Message) (any, error, error) {
	desc := "`!pb play gamemode #rounds`\n" +
		// "`!pb play [f/m/t] [#]`\n" +
		"`!pb play [f] [#]`\n" +
		"\n" +
		"The **maximum** rounds you can play is 15.\n" +
		"\n" +
		"Example:\n" +
		"*To play 10 rounds in Free-For-All.*\n" +
		"`!pb play f 10`"

	help := &dg.MessageEmbed{
		Title:       "⚙ **Choose your settings:** ⚙",
		Description: desc,
		Color:       embedColor,
	}

	if len(m.Command.Runtime.Args) < 2 {
		return help, nil, nil
	}

	mode := m.Command.Runtime.Args[0]
	if !(mode == "f" || mode == "m" || mode == "t") {
		return help, nil, nil
	}

	rounds, err := strconv.Atoi(m.Command.Runtime.Args[1])
	if err != nil {
		return help, nil, nil
	}
	if rounds < 1 || rounds > 15 {
		return help, nil, nil
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

	switch mode {
	case "f":
		return runPlayF(m.Channel.ID, here, rounds)
	}

	return nil, nil, nil
}

func runPlayF(channel string, here int64, rounds int) (*dg.MessageEmbed, error, error) {
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
			name = answer.Author.DisplayName
			game.Point(here, answer.Author)
		}

		resp := generateAnswer(r, name, poster, answers, r == rounds)
		write(channel, resp)
	}

	write(channel, generateScorecard(game.Scores(here)))
	game.Playing(here, false)
	return nil, nil, core.ErrSilence
}
