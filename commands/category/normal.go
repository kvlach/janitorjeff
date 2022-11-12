package category

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return frontends.Twitch
}

func (normal) Names() []string {
	return []string{
		"category",
		"game",
	}
}

func (normal) Description() string {
	return "Change or see the current category."
}

func (normal) UsageArgs() string {
	return "[category]"
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
	switch m.Frontend {
	case frontends.Twitch:
		return c.twitch(m)
	default:
		panic("this should never happen")
	}
}

func (normal) twitch(m *core.Message) (string, error, error) {
	h, err := m.Client.(*twitch.Twitch).Helix()
	if err != nil {
		return "", nil, err
	}

	if len(m.Command.Args) == 0 {
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
