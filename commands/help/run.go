package help

import (
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func run(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return run_Discord(m)
	default:
		return run_Text(m)
	}
}

func run_Text(m *core.Message) (string, error, error) {
	cmd, aliases, err := run_Core(m)
	if err != nil {
		return "", nil, err
	}

	help := fmt.Sprintf("Usage: %s.", cmd.Usage())

	if cmd.Static.Description != "" {
		help += " " + cmd.Static.Description
	}

	if len(aliases) > 0 {
		help += fmt.Sprintf(" Aliases: %s.", strings.Join(aliases, ", "))
	}

	log.Debug().Str("help", help).Send()

	return m.ReplyText(help), nil, nil
}

func run_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	cmd, aliases, err := run_Core(m)
	if err != nil {
		return nil, nil, err
	}

	desc := ""

	if cmd.Static.Description != "" {
		desc += fmt.Sprintf("*%s*", cmd.Static.Description)
	}

	if len(aliases) > 0 {
		var base string
		if len(cmd.Runtime.Name) == 1 {
			base = cmd.Runtime.Prefix
		} else {
			cmdBase := strings.Join(cmd.Runtime.Name[:len(cmd.Runtime.Name)-1], " ")
			base = fmt.Sprintf("%s%s ", cmd.Runtime.Prefix, cmdBase)
		}

		for i := range aliases {
			aliases[i] = fmt.Sprintf("- `%s%s`", base, aliases[i])
		}
		desc += fmt.Sprintf("\n\nAliases:\n%s", strings.Join(aliases, "\n"))
	}

	embed := &dg.MessageEmbed{
		Title:       fmt.Sprintf("Usage: `%s`", cmd.Usage()),
		Description: desc,
	}

	log.Debug().Interface("embed", embed).Send()

	return embed, nil, nil
}

func run_Core(m *core.Message) (*core.Command, []string, error) {
	cmdStatic, index, err := core.Globals.Commands.Normal.MatchCommand(m.Command.Runtime.Args)
	if err != nil {
		return nil, nil, err
	}

	cmd := &core.Command{
		Static: cmdStatic,
		Runtime: &core.CommandRuntime{
			Name:   m.Command.Runtime.Args[:index+1],
			Prefix: m.Command.Runtime.Prefix,
		},
	}

	cmdName := cmd.Runtime.Name[len(cmd.Runtime.Name)-1]
	aliases := make([]string, 0, len(cmdStatic.Names)-1)

	for _, name := range cmdStatic.Names {
		if name != cmdName {
			aliases = append(aliases, name)
		}
	}

	log.Debug().
		Strs("names", cmd.Static.Names).
		Strs("aliases", aliases).
		Str("current-name", cmdName).
		Msg("filtered out command aliases")

	return cmd, aliases, nil
}
