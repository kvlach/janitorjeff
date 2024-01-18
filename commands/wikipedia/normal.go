package wikipedia

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

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
	return "<query>"
}

func (normal) Category() core.CommandCategory {
	return core.CommandCategoryServices
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
	return nil
}

func (c normal) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normal) discord(m *core.Message) (any, core.Urr, error) {
	res, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	// simply send the url with no embed and let discord create an embed
	if urr == nil {
		return res.Canonicalurl, nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, res),
	}
	return embed, urr, nil
}

func (c normal) text(m *core.Message) (string, core.Urr, error) {
	res, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr, res), urr, nil
}

func (normal) fmt(urr core.Urr, res page) string {
	switch urr {
	case nil:
		return res.Canonicalurl
	case UrrNoResult:
		return "Couldn't find anything."
	default:
		return fmt.Sprint(urr)
	}
}

func (normal) core(m *core.Message) (page, core.Urr, error) {
	return Search(m.RawArgs(0))
}
