package core

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// The command types.
const (
	// A simplified command with might not give full control over something but
	// it has a very easy to use API.
	Normal = 1 << iota

	// The full command and usually consists of many subcommands which makes it
	// less intuitive for the average person.
	Advanced

	// Bot admin only command used to perform actions like setting an arbitrary
	// person's options, etc.
	Admin

	All = Normal | Advanced | Admin
)

// There's 2 parts to a command. The static part which includes things like the
// description, the list of all the aliases, etc. and the runtime part which
// includes things like the prefix used, the arguments passed, etc.

// The struct used to declare commands.
type CommandStatic struct {
	// All the aliases a command has. The first item in the list is considered
	// the main name and so should be the simplest and most intuitive one for
	// the average person. For example if it's a delete subcommand the first
	// alias should be "delete" instead of "del" or "rm".
	Names []string

	// A short description of what the command does.
	Description string

	// Usage arguments. Should follow this format:
	// - <required>
	// - [optional]
	// - (literal-string) or (many | literals)
	UsageArgs string

	// The frontends where this command will be available at. This *must* be set
	// otherwise the command will never get matched.
	Frontends int

	// The function that is called to run the command.
	Run func(*Message) (any, error, error)

	// This is executed during bot startup. Should be used to set things up
	// necessary for the command, for example DB schemas.
	Init func() error

	// The command's sub-commands.
	Children Commands

	// A command's parent, this is automatically set during bot startup.
	Parent *CommandStatic
}

// Formats a static command into something that can be shown to a user.
// Generally used in help messages to point the user to a specific command in
// order to avoid hardcoding it. Returns the command in the following format:
//
//	<prefix><command> [sub-command...] <usage-args>
//
// For example: !command delete <command>
func (cmd *CommandStatic) Format(prefix string) string {
	path := []string{}
	for cmd.Parent != nil {
		path = append([]string{cmd.Names[0]}, path...)
		cmd = cmd.Parent
	}
	path = append([]string{cmd.Names[0]}, path...)

	var args string
	if cmd.UsageArgs != "" {
		args = " " + cmd.UsageArgs
	}

	return fmt.Sprintf("%s%s%s", prefix, strings.Join(path, " "), args)
}

// A command's runtime information.
type CommandRuntime struct {
	// Includes all the sub-commands e.g. ["prefix", "add"], so that we can
	// know which alias is being used in order to display accurate help
	// messages.
	Name []string

	// The arguments passed, includes everything that's not part of the
	// command's name.
	Args []string

	// The prefix used when the command was called.
	Prefix string
}

type Command struct {
	Type    int
	Static  *CommandStatic
	Runtime *CommandRuntime
}

func (cmd *Command) Usage() string {
	cmdName := strings.Join(cmd.Runtime.Name, " ")
	usage := fmt.Sprintf("%s%s", cmd.Runtime.Prefix, cmdName)

	if cmd.Static.UsageArgs != "" {
		usage = fmt.Sprintf("%s %s", usage, cmd.Static.UsageArgs)
	}

	return usage
}

type Commands []*CommandStatic

func (cmds *Commands) matchCommand(frontend int, cmdName string) (*CommandStatic, error) {
	if cmdName == "" {
		return nil, fmt.Errorf("no command provided")
	}

	cmdName = strings.ToLower(cmdName)

	for _, c := range *cmds {
		if c.Frontends&frontend == 0 {
			continue
		}

		for _, a := range c.Names {
			if a == cmdName {
				log.Debug().Str("command", cmdName).Msg("matched command")
				return c, nil
			}
		}
	}

	return nil, fmt.Errorf("command '%s' not found", cmdName)
}

// Given a list of arguments match the appropriate command. The arguments don't
// have to correspond to a command exactly. For example,
//
// `args = [prefix add abc]`.
//
// In this case the prefix's subcommand `add` will be matched and returned.
// Alongside it the index of the last valid command will be returned (in this
// case the index of "add", which is 1). This makes it easy to know which
// aliases where used by the user when invoking a command.
func (cmds *Commands) MatchCommand(frontend int, args []string) (*CommandStatic, int, error) {
	log.Debug().Strs("args", args).Msg("trying to match command")

	index := 0

	cmd, err := cmds.matchCommand(frontend, args[index])
	if err != nil {
		return nil, -1, err
	}

	for _, c := range args[1:] {
		tmp, err := cmd.Children.matchCommand(frontend, c)
		if err != nil {
			return cmd, index, nil
		}
		index++
		cmd = tmp
	}

	return cmd, index, nil
}
