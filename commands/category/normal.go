package category

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"category",
		"game",
	},
	Description: "Change or see the current category.",
	UsageArgs:   "[category]",
	Frontends:   frontends.Twitch,
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Twitch:
		return normalRunTwitch(m)
	default:
		panic("this should never happen")
	}
}

func normalRunTwitch(m *core.Message) (string, error, error) {
	helix := m.Client.(*twitch.Twitch).Helix

	if len(m.Command.Runtime.Args) == 0 {
		g, err := helix.GetGameName(m.Channel.ID)
		return g, nil, err
	}

	g, err := helix.SetGame(m.Channel.ID, m.RawArgs(0))

	switch err {
	case nil:
		return fmt.Sprintf("Category set to: %s", g), nil, nil
	case twitch.ErrNoResults:
		return "Couldn't find the game, did you type the name correctly?", nil, nil
	default:
		return "", nil, err
	}
}
