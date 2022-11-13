package title

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

func (normal) Permitted(*core.Message) bool {
	return true
}

func (normal) Names() []string {
	return []string{
		"title",
	}
}

func (normal) Description() string {
	return "Change or see the current title."
}

func (normal) UsageArgs() string {
	return "[title]"
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
