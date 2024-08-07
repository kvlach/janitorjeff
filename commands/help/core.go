package help

import (
	"fmt"
	"strings"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// The help command is slightly unique in the fact that its functionality
// across all its command types is exactly the same with the only difference
// being the list of commands from which it matches (e.g. if it is called with
// a normal prefix it will match normal commands). This means that almost all
// of its functionality is implemented in the core, including the rendering.

var cmdNames = []string{
	"help",
}

const (
	cmdDescription = "Shows a help message for the specified %s command."
	cmdUsageArgs   = "<command...>"
)

var UrrCommandNotFound = core.UrrNew("Command could not be found.")

func runCore(t core.CommandType, m *core.EventMessage, args []string, prefix string) (*core.Command, []string, []string, core.Urr) {
	cmdStatic, index, err := core.Commands.Match(t, m, args)
	if err != nil {
		return nil, nil, nil, UrrCommandNotFound
	}

	cmd := &core.Command{
		CommandStatic: cmdStatic,
		CommandRuntime: core.CommandRuntime{
			Path:   args[:index+1],
			Prefix: prefix,
		},
	}

	// Convert to lower case in order to correctly match the command's name in
	// the list of aliases (since they are all lowercase, and the user could
	// have typed the command otherwise)
	cmdName := strings.ToLower(cmd.Path[len(cmd.Path)-1])

	aliases := make([]string, 0, len(cmdStatic.Names())-1)

	for _, name := range cmdStatic.Names() {
		if name != cmdName {
			aliases = append(aliases, name)
		}
	}

	log.Debug().
		Strs("aliases", aliases).
		Str("name", cmdName).
		Msg("filtered out command aliases")

	return cmd, aliases, cmdStatic.Examples(), nil
}

func renderText(cmd *core.Command, aliases []string) string {
	var help strings.Builder
	fmt.Fprintf(&help, "Usage: %s.", cmd.Usage())

	if cmd.Description() != "" {
		help.WriteString(" " + cmd.Description())
	}

	if len(aliases) > 0 {
		fmt.Fprintf(&help, " Aliases: %s.", strings.Join(aliases, ", "))
	}

	return help.String()
}

func renderDiscord(cmd *core.Command, aliases []string, examples []string) *dg.MessageEmbed {
	var desc strings.Builder

	if cmd.Description() != "" {
		fmt.Fprintf(&desc, "*%s*", cmd.Description())
	}

	if len(examples) > 0 {
		base := fmt.Sprintf("%s%s ", cmd.Prefix, strings.Join(cmd.Path, " "))
		for i := range examples {
			if len(examples[i]) == 0 {
				// trim space that is present so no arg follows
				examples[i] = fmt.Sprintf("- `%s`", base[:len(base)-1])
			} else {
				examples[i] = fmt.Sprintf("- `%s%s`", base, examples[i])
			}
		}
		fmt.Fprintf(&desc, "\n\nExamples:\n%s", strings.Join(examples, "\n"))
	}

	if len(aliases) > 0 {
		var base string
		if len(cmd.Path) == 1 {
			base = cmd.Prefix
		} else {
			cmdBase := strings.Join(cmd.Path[:len(cmd.Path)-1], " ")
			base = fmt.Sprintf("%s%s ", cmd.Prefix, cmdBase)
		}

		for i := range aliases {
			aliases[i] = fmt.Sprintf("- `%s%s`", base, aliases[i])
		}
		fmt.Fprintf(&desc, "\n\nAliases:\n%s", strings.Join(aliases, "\n"))
	}

	embed := &dg.MessageEmbed{
		Title:       fmt.Sprintf("Usage: `%s`", cmd.Usage()),
		Description: desc.String(),
	}

	return embed
}

func runDiscord(t core.CommandType, m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	cmd, aliases, examples, urr := runCore(t, m, m.Command.Args, m.Command.Prefix)
	if urr != nil {
		return &dg.MessageEmbed{Description: fmt.Sprint(urr)}, urr, nil
	}
	return renderDiscord(cmd, aliases, examples), nil, nil
}

func runText(t core.CommandType, m *core.EventMessage) (string, core.Urr, error) {
	cmd, aliases, _, urr := runCore(t, m, m.Command.Args, m.Command.Prefix)
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}
	return renderText(cmd, aliases), nil, nil
}

func run(t core.CommandType, m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return runDiscord(t, m)
	default:
		return runText(t, m)
	}
}
