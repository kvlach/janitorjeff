package title

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"title",
	},
	Description: "Change or see the current title.",
	UsageArgs:   "[title]",
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
		t, err := h.GetTitle(m.Channel.ID)
		return t, nil, err
	}

	title := m.RawArgs(0)

	usrErr, err := h.SetTitle(m.Channel.ID, title)
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("Title set to: %s", title), nil, nil
}
