package wikipedia

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

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
		"wikipedia",
		"wiki",
	}
}

func (normal) Description() string {
	return "Search something on wikipedia."
}

func (normal) UsageArgs() string {
	return "<query...>"
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

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normal) discord(m *core.Message) (any, error, error) {
	res, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	// simply send the url with no embed and let discord create an embed
	if usrErr == nil {
		return res.Canonicalurl, nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, res),
	}
	return embed, usrErr, nil
}

func (c normal) text(m *core.Message) (string, error, error) {
	res, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr, res), usrErr, nil
}

func (normal) err(usrErr error, res page) string {
	switch usrErr {
	case nil:
		return res.Canonicalurl
	case errNoResult:
		return "Couldn't find anything."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (normal) core(m *core.Message) (page, error, error) {
	return search(m.RawArgs(0))
}
