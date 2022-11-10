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
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}

	if len(m.Command.Runtime.Args) == 0 {
		g, err := h.GetGameName(m.Channel.ID)
		return g, nil, err
	}

	g, usrErr, err := h.SetGame(m.Channel.ID, m.RawArgs(0))

	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}

	switch err {
	case nil:
		return fmt.Sprintf("Category set to: %s", g), nil, nil
	case twitch.ErrNoResults:
		return "Couldn't find the category, did you type the name correctly?", nil, nil
	default:
		return "", nil, err
	}
}
