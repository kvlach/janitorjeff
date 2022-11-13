package core

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type CommandType int

// The command types.
const (
	// A simplified command with might not give full control over something but
	// it has a very easy to use API.
	Normal CommandType = 1 << iota

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

// A command needs to implement this interface.
type CommandStatic interface {
	// The command's type.
	Type() CommandType

	// The frontends where this command will be available at.
	Frontends() int

	// Any other checks required for a command to be executed. Returns true if
	// the command is allowed to be executed. Usually is just a mod/admin check.
	Permitted(m *Message) bool

	// All the aliases a command has. The first item in the list is considered
	// the main name and so should be the simplest and most intuitive one for
	// the average person. For example if it's a delete subcommand the first
	// alias should be "delete" instead of "del" or "rm".
	Names() []string

	// A short description of what the command does.
	Description() string

	// Usage arguments. Should follow this format:
	// - <required>
	// - [optional]
	// - (literal-string) or (many | literals)
	UsageArgs() string

	// A command's parent, this is automatically set during bot startup.
	Parent() CommandStatic

	// The command's sub-commands.
	Children() CommandsStatic

	// This is executed during bot startup. Should be used to set things up
	// necessary for the command, for example DB schemas.
	Init() error

	// The function that is called to run the command.
	Run(m *Message) (resp any, usrErr error, err error)
}

// Formats a static command into something that can be shown to a user.
// Generally used in help messages to point the user to a specific command in
// order to avoid hardcoding it. Returns the command in the following format:
//
//	<prefix><command> [sub-command...] <usage-args>
//
// For example: !command delete <command>
func Format(cmd CommandStatic, prefix string) string {
	var args string
	if cmd.UsageArgs() != "" {
		args = " " + cmd.UsageArgs()
	}

	path := []string{}
	for cmd.Parent() != nil {
		path = append([]string{cmd.Names()[0]}, path...)
		cmd = cmd.Parent()
	}
	path = append([]string{cmd.Names()[0]}, path...)

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
	CommandStatic
	CommandRuntime
}

func (cmd *Command) Usage() string {
	usage := cmd.Prefix + strings.Join(cmd.Name, " ")
	if cmd.UsageArgs() != "" {
		usage += " " + cmd.UsageArgs()
	}
	return usage
}

type CommandsStatic []CommandStatic

func (cmds CommandsStatic) match(frontend int, t CommandType, name string) (CommandStatic, error) {
	name = strings.ToLower(name)

	for _, c := range cmds {
		if c.Frontends()&frontend == 0 {
			continue
		}

		if c.Type() != t {
			continue
		}

		for _, n := range c.Names() {
			if name == n {
				log.Debug().Str("command", name).Msg("matched command")
				return c, nil
			}
		}
	}

	return nil, fmt.Errorf("command '%s' not found", name)
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
func (cmds *CommandsStatic) Match(t CommandType, frontend int, args []string) (CommandStatic, int, error) {
	log.Debug().Strs("args", args).Msg("trying to match command")

	index := 0

	cmd, err := cmds.match(frontend, t, args[0])
	if err != nil {
		return nil, -1, err
	}

	for _, name := range args[1:] {
		tmp, err := cmd.Children().match(frontend, t, name)
		if err != nil {
			return cmd, index, nil
		}
		index++
		cmd = tmp
	}

	return cmd, index, nil
}
