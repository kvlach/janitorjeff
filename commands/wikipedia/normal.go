package wikipedia

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"wikipedia",
		"wiki",
	},
	Description: "Search something on wikipedia.",
	UsageArgs:   "<query...>",
	Frontends:   frontends.All,
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return normalRunDiscord(m)
	default:
		return normalRunText(m)
	}
}

func normalRunDiscord(m *core.Message) (any, error, error) {
	res, usrErr, err := normalRunCore(m)
	if err != nil {
		return nil, nil, err
	}

	// simply send the url with no embed and let discord create an embed
	if usrErr == nil {
		return res.Canonicalurl, nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: normalRunErr(usrErr, res),
	}
	return embed, usrErr, nil
}

func normalRunText(m *core.Message) (string, error, error) {
	res, usrErr, err := normalRunCore(m)
	if err != nil {
		return "", nil, err
	}
	return normalRunErr(usrErr, res), usrErr, nil
}

func normalRunErr(usrErr error, res page) string {
	switch usrErr {
	case nil:
		return res.Canonicalurl
	case errNoResult:
		return "Couldn't find anything."
	default:
		return fmt.Sprint(usrErr)
	}
}

func normalRunCore(m *core.Message) (page, error, error) {
	return search(m.RawArgs(0))
}
