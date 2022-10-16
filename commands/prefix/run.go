package prefix

import (
	"errors"
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/commands/command"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	errMissingArgument     = errors.New("missing argument")
	errExists              = errors.New("prefix exists already")
	errNotFound            = errors.New("prefix not found")
	errOneLeft             = errors.New("only one prefix left")
	errCustomCommandExists = errors.New("if this prefix is added then there will be a collision with a custom command")
)

func run(m *core.Message) (interface{}, error, error) {
	return m.ReplyUsage(), errMissingArgument, nil
}

func runAdd(m *core.Message) (interface{}, error, error) {
	switch m.Type {
	case core.Discord:
		return runAdd_Discord(m)
	default:
		return runAdd_Text(m)
	}
}

func runAdd_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefix, collision, usrErr, err := runAdd_Core(m)
	if err != nil {
		return nil, usrErr, err
	}

	prefix = discord.PlaceInBackticks(prefix)
	collision = discord.PlaceInBackticks(collision)

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(*dg.MessageEmbed), usrErr, nil
	}

	embed := &dg.MessageEmbed{
		Description: runAdd_Err(usrErr, prefix, collision),
	}

	return embed, usrErr, nil
}

func runAdd_Text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefix, collision, usrErr, err := runAdd_Core(m)
	if err != nil {
		return "", usrErr, err
	}

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(string), usrErr, nil
	}

	return runAdd_Err(usrErr, prefix, collision), usrErr, nil
}

func runAdd_Err(err error, prefix, collision string) string {
	switch err {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case errExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case errCustomCommandExists:
		return fmt.Sprintf("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.", prefix, collision)
	default:
		return "Something went wrong..."
	}
}

func runAdd_Core(m *core.Message) (string, string, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return "", "", errMissingArgument, nil
	}
	prefix := m.Command.Runtime.Args[0]

	scope, err := m.ScopePlace()
	if err != nil {
		return prefix, "", nil, err
	}
	log.Debug().Int64("scope", scope).Send()

	prefixes, scopeExists, err := m.ScopePrefixes()
	if err != nil {
		return prefix, "", nil, err
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
			if p.Type != core.Normal {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Normal); err != nil {
				return prefix, "", nil, err
			}
		}
	}

	exists, err := dbPrefixExists(prefix, scope, core.Normal)
	if err != nil {
		return prefix, "", nil, err
	}
	if exists {
		return prefix, "", errExists, nil
	}

	collision, err := customCommandCollision(m, prefix)
	if err != nil {
		return prefix, "", nil, err
	}
	if collision != "" {
		return prefix, collision, errCustomCommandExists, nil
	}

	err = dbAdd(prefix, scope, core.Normal)
	return prefix, "", nil, err
}

// if the prefix changes after a custom command has been added it's
// possible that a collision maybe be created
//
// for example:
// !prefix reset
// !cmd add .prefix test // this works because . is not a valid prefix atm
// !prefix add .
// .prefix // both trigger
func customCommandCollision(m *core.Message, prefix string) (string, error) {
	triggers, err := command.RunList_Core(m)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		t = strings.TrimPrefix(t, prefix)
		_, _, err := core.Globals.Commands.Normal.MatchCommand([]string{t})
		if err == nil {
			return prefix + t, nil
		}
	}

	return "", nil
}

func runDelete(m *core.Message) (interface{}, error, error) {
	// TODO: Provide a way to delete the empty string prefix in DMs
	switch m.Type {
	case core.Discord:
		return runDelete_Discord(m)
	default:
		return runDelete_Text(m)
	}
}

func runDelete_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefix, usrErr, err := runDelete_Core(m)
	if err != nil {
		return nil, usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(*dg.MessageEmbed), usrErr, nil
	case errOneLeft:
		resetCommand = cmdReset.Format(m.Command.Runtime.Prefix)
		resetCommand = discord.PlaceInBackticks(resetCommand)
	}

	prefix = discord.PlaceInBackticks(prefix)

	embed := &dg.MessageEmbed{
		Description: runDelete_Err(usrErr, m, prefix, resetCommand),
	}

	return embed, usrErr, nil
}

func runDelete_Text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefix, usrErr, err := runDelete_Core(m)
	if err != nil {
		return "", usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case errMissingArgument:
		return m.ReplyUsage().(string), usrErr, nil
	case errOneLeft:
		resetCommand = cmdReset.Format(m.Command.Runtime.Prefix)
	}

	return runDelete_Err(usrErr, m, prefix, resetCommand), usrErr, nil
}

func runDelete_Err(err error, m *core.Message, prefix, resetCommand string) string {
	switch err {
	case nil:
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

func runDelete_Core(m *core.Message) (string, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return "", errMissingArgument, nil
	}
	prefix := m.Command.Runtime.Args[0]

	scope, err := m.ScopePlace()
	if err != nil {
		return prefix, nil, err
	}

	log.Debug().
		Str("prefix", prefix).
		Int64("scope", scope).
		Send()

	prefixes, scopeExists, err := m.ScopePrefixes()
	if err != nil {
		return prefix, nil, err
	}
	var exists bool
	for _, p := range prefixes {
		if p.Type != core.Normal {
			continue
		}

		if p.Prefix == prefix {
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
			if p.Type != core.Normal {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Normal); err != nil {
				return prefix, nil, err
			}
		}
	}

	return prefix, nil, dbDel(prefix, scope)
}

func runList(m *core.Message) (interface{}, error, error) {
	switch m.Type {
	case core.Discord:
		return runList_Discord(m)
	default:
		return runList_Text(m)
	}
}

func runList_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefixes, err := runList_Core(m)
	if err != nil {
		return nil, nil, err
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

	return embed, nil, nil
}

func runList_Text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefixes, err := runList_Core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes: %s", strings.Join(prefixes, " ")), nil, nil
}

func runList_Core(m *core.Message) ([]string, error) {
	prefixes, _, err := m.ScopePrefixes()

	normal := []string{}
	for _, p := range prefixes {
		if p.Type == core.Normal {
			normal = append(normal, p.Prefix)
		}
	}

	return normal, err
}

func runReset(m *core.Message) (interface{}, error, error) {
	switch m.Type {
	case core.Discord:
		return runReset_Discord(m)
	default:
		return runReset_Text(m)
	}
}

func runReset_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	listCmd, err := runReset_Core(m)
	if err != nil {
		return nil, nil, err
	}
	listCmd = discord.PlaceInBackticks(listCmd)

	embed := &dg.MessageEmbed{
		Description: fmt.Sprintf("Prefixes have been reset.\nTo view the list of the currently available prefixes run: %s", listCmd),
	}

	return embed, nil, nil
}

func runReset_Text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	listCmd, err := runReset_Core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes have been reset. To view the list of the currently available prefixes run: %s", listCmd), nil, nil
}

func runReset_Core(m *core.Message) (string, error) {
	scope, err := m.ScopePlace()
	if err != nil {
		return "", err
	}

	err = dbReset(scope)
	if err != nil {
		return "", err
	}

	var prefix string
	for _, p := range core.Globals.Prefixes.Others {
		if p.Type == core.Normal {
			prefix = p.Prefix
			break
		}
	}

	// can't just use the prefix that was used to invoke this command because
	// it might not be valid for this scope since a reset just happened
	listCmd := cmdList.Format(prefix)

	return listCmd, nil
}
