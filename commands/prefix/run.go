package prefix

import (
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	noErr = -1

	errMissingArgument = iota
	errExists
	errNotFound
	errOneLeft
)

func run(m *core.Message) (interface{}, error) {
	return m.ReplyUsage(), nil
}

func runAdd(m *core.Message) (interface{}, error) {
	switch m.Type {
	case core.Discord:
		return runAdd_Discord(m)
	default:
		return runAdd_Text(m)
	}
}

func runAdd_Discord(m *core.Message) (*dg.MessageEmbed, error) {
	log.Debug().Msg("running discord renderer")

	prefix, usrErr, err := runAdd_Core(m)
	if err != nil {
		return nil, err
	}

	prefix = discord.PlaceInBackticks(prefix)

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(*dg.MessageEmbed), nil
	}

	embed := &dg.MessageEmbed{
		Description: runAdd_Err(usrErr, prefix),
	}

	return embed, nil
}

func runAdd_Text(m *core.Message) (string, error) {
	log.Debug().Msg("running plain text renderer")

	prefix, usrErr, err := runAdd_Core(m)
	if err != nil {
		return "", err
	}

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(string), nil
	}

	return runAdd_Err(usrErr, prefix), nil
}

func runAdd_Err(err int, prefix string) string {
	switch err {
	case noErr:
		return fmt.Sprintf("Added prefix %s", prefix)
	case errExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	default:
		return "Something went wrong..."
	}
}

func runAdd_Core(m *core.Message) (string, int, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return "", errMissingArgument, nil
	}
	prefix := m.Command.Runtime.Args[0]

	scope, err := m.Scope()
	if err != nil {
		return prefix, noErr, err
	}
	log.Debug().Int64("scope", scope).Send()

	prefixes, scopeExists, err := m.ScopePrefixes()
	if err != nil {
		return prefix, noErr, err
	}

	log.Debug().
		Bool("scopeExists", scopeExists).
		Msg("checked if custom prefixes have been added before for this scope")

	// Only add default prefixes if they've never been added before, this
	// prevents situations were the default prefixes change and they sneakily
	// get added without the user realizing.
	if !scopeExists {
		log.Debug().Msg("adding default prefixes to scope prefixes")

		for _, p := range prefixes {
			if err = dbAdd(p, scope); err != nil {
				return prefix, noErr, err
			}
		}
	}

	exists, err := dbExists(prefix, scope)
	if err != nil {
		return prefix, noErr, err
	}
	if exists {
		return prefix, errExists, nil
	}

	err = dbAdd(prefix, scope)
	return prefix, noErr, err
}

func runDelete(m *core.Message) (interface{}, error) {
	// TODO: Provide a way to delete the empty string prefix in DMs
	switch m.Type {
	case core.Discord:
		return runDelete_Discord(m)
	default:
		return runDelete_Text(m)
	}
}

func runDelete_Discord(m *core.Message) (*dg.MessageEmbed, error) {
	log.Debug().Msg("running discord renderer")

	prefix, usrErr, err := runDelete_Core(m)
	if err != nil {
		return nil, err
	}

	resetCommand := ""

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(*dg.MessageEmbed), nil
	case errOneLeft:
		resetCommand = cmdReset.Format(m.Command.Runtime.Prefix)
		resetCommand = discord.PlaceInBackticks(resetCommand)
	}

	prefix = discord.PlaceInBackticks(prefix)

	embed := &dg.MessageEmbed{
		Description: runDelete_Err(usrErr, m, prefix, resetCommand),
	}

	return embed, nil
}

func runDelete_Text(m *core.Message) (string, error) {
	log.Debug().Msg("running plain text renderer")

	prefix, usrErr, err := runDelete_Core(m)
	if err != nil {
		return "", err
	}

	resetCommand := ""

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(string), nil
	case errOneLeft:
		resetCommand = cmdReset.Format(m.Command.Runtime.Prefix)
	}

	return runDelete_Err(usrErr, m, prefix, resetCommand), nil
}

func runDelete_Err(err int, m *core.Message, prefix, resetCommand string) string {
	switch err {
	case noErr:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case errNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case errOneLeft:
		return fmt.Sprintf("Can't delete, %s is the only prefix left.\n", prefix) +
			fmt.Sprintf("If you wish to reset to the default prefixes run: %s", resetCommand)
	default:
		return "Something went wrong..."
	}
}

func runDelete_Core(m *core.Message) (string, int, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return "", errMissingArgument, nil
	}
	prefix := m.Command.Runtime.Args[0]

	scope, err := m.Scope()
	if err != nil {
		return prefix, noErr, err
	}

	log.Debug().
		Str("prefix", prefix).
		Int64("scope", scope).
		Send()

	prefixes, scopeExists, err := m.ScopePrefixes()
	if err != nil {
		return prefix, noErr, err
	}
	var exists bool
	for _, p := range prefixes {
		if p == prefix {
			exists = true
		}
	}

	if !exists {
		return prefix, errNotFound, nil
	}

	if len(prefixes) == 1 {
		return prefix, errOneLeft, nil
	}

	// If the scope doesn't exist then the default prefixes are being used and
	// they are not present in the DB. So if the user tries to delete one
	// nothing will happen. So we first add them all to the DB.
	if !scopeExists {
		for _, p := range prefixes {
			if err = dbAdd(p, scope); err != nil {
				return prefix, noErr, err
			}
		}
	}

	return prefix, noErr, dbDel(prefix, scope)
}

func runList(m *core.Message) (interface{}, error) {
	switch m.Type {
	case core.Discord:
		return runList_Discord(m)
	default:
		return runList_Text(m)
	}
}

func runList_Discord(m *core.Message) (*dg.MessageEmbed, error) {
	log.Debug().Msg("running discord renderer")

	prefixes, err := runList_Core(m)
	if err != nil {
		return nil, err
	}

	for i, p := range prefixes {
		// TODO: Add a warning message saying that the prefix can not be
		// displayed correctly on android devices
		prefixes[i] = discord.PlaceInBackticks(p + "command")
	}

	embed := &dg.MessageEmbed{
		Title:       "Prefixes",
		Description: strings.Join(prefixes, "\n"),
	}

	return embed, nil
}

func runList_Text(m *core.Message) (string, error) {
	log.Debug().Msg("running plain text renderer")

	prefixes, err := runList_Core(m)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Prefixes: %s", strings.Join(prefixes, " ")), nil
}

func runList_Core(m *core.Message) ([]string, error) {
	prefixes, _, err := m.ScopePrefixes()
	return prefixes, err
}

func runReset(m *core.Message) (interface{}, error) {
	switch m.Type {
	case core.Discord:
		return runReset_Discord(m)
	default:
		return runReset_Text(m)
	}
}

func runReset_Discord(m *core.Message) (*dg.MessageEmbed, error) {
	log.Debug().Msg("running discord renderer")

	listCmd, err := runReset_Core(m)
	if err != nil {
		return nil, err
	}
	listCmd = discord.PlaceInBackticks(listCmd)

	embed := &dg.MessageEmbed{
		Description: fmt.Sprintf("Prefixes have been reset.\nTo view the list of the currently available prefixes run: %s", listCmd),
	}

	return embed, nil
}

func runReset_Text(m *core.Message) (string, error) {
	log.Debug().Msg("running plain text renderer")

	listCmd, err := runReset_Core(m)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Prefixes have been reset. To view the list of the currently available prefixes run: %s", listCmd), nil
}

func runReset_Core(m *core.Message) (string, error) {
	scope, err := m.Scope()
	if err != nil {
		return "", err
	}

	err = dbReset(scope)
	if err != nil {
		return "", err
	}

	// can't just use the prefix that was used to invoke this command because
	// it might not be valid for this scope since a reset just happened
	listCmd := cmdList.Format(core.Globals.Prefixes()[0])

	return listCmd, nil
}
