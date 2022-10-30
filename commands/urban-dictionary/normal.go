package urban_dictionary

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"ud",
	},
	Description: "Search a term on urban dictionary.",
	UsageArgs:   "<term...>",
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return normalRunDiscord(m)
	default:
		return normalRunText(m)
	}
}

func normalRunDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	def, err := normalRunCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Title:       "UrbanDictionary definition for " + def.Word,
		URL:         def.Permalink,
		Description: def.Definition,
		Fields: []*dg.MessageEmbedField{
			{
				Name:  "Example",
				Value: def.Example,
			},
		},
		Footer: &dg.MessageEmbedFooter{
			Text: fmt.Sprintf("Submitter: %s | Thumbs up: %d | Thumbs down: %d", def.Author, def.ThumbsUp, def.ThumbsDown),
		},
	}

	return embed, nil, nil
}

func normalRunText(m *core.Message) (string, error, error) {
	return "", nil, nil
}

func normalRunCore(m *core.Message) (definition, error) {
	term := m.RawArgs(0)
	return search(term)
}
