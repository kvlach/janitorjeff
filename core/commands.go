package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

var Commands CommandsStatic

type CommandType int

// The command types.
const (
	// Normal represents a simplified command which might not give full control
	// over something, but it has a very easy to use API.
	Normal CommandType = 1 << iota

	// Advanced represents the full command and usually consists of many
	// subcommands which makes it less intuitive for the average person.
	Advanced

	// Admin represents a bot-admin-only command used to perform actions like
	// setting an arbitrary person's info, etc.
	Admin

	All = Normal | Advanced | Admin
)

type CommandCategory string

const (
	CommandCategoryGames      = "Games"
	CommandCategoryModerators = "Moderators"
	CommandCategoryOther      = "Other"
	CommandCategoryServices   = "Services"
)

// There are 2 parts to a command: the static part which includes things like
// the description, the list of all the aliases, etc. and the runtime part which
// includes things like the prefix used, the arguments passed, etc.

// CommandStatic is the interface used to implement commands.
type CommandStatic interface {
	// Type returns the command's type.
	Type() CommandType

	// Permitted will perform checks required for a command to be executed.
	// Returns true if the command is allowed to be executed. Usually used to
	// check a user's permissions or to restrict a command to specific
	// frontends.
	Permitted(m *EventMessage) bool

	// Names return a list of all the aliases a command has. The first item in
	// the list is considered the main name and so should be the simplest and
	// most intuitive one for the average person. For example, if it's a deleted
	// subcommand, the first alias should be "delete" instead of "del" or "rm."
	Names() []string

	// Description will return a short description of what the command does.
	Description() string

	// UsageArgs will return the usage arguments. Should follow this format:
	// - <required>
	// - [optional]
	// - (literal-string) or (many | literals)
	UsageArgs() string

	// Category returns the general category the command belongs to. Mainly
	// to make displaying all the commands easier and less overwhelming (as
	// they are split up instead of having them all in a giant list).
	Category() CommandCategory

	// Examples returns a list of strings with example usages of the command.
	// Only the arguments passed should be included, not the prefix and the
	// chain of command names.
	Examples() []string

	// Parent returns a command's parent, returns nil if there is no parent.
	Parent() CommandStatic

	// Children returns the command's sub-commands, returns nil if there are no
	// sub-commands.
	Children() CommandsStatic

	// Init is executed during bot startup. Should be used to set things up
	// necessary for the command, for example, DB schemas.
	Init() error

	// Run is the function that is called to run the command.
	Run(m *EventMessage) (resp any, urr Urr, err error)
}

// Format will return a string representation of the given command in a format
// that can be shown to a user. Generally used in help messages to point the
// user to a specific command in order to avoid hard-coding it. Returns the
// command in the following format:
//
//	<prefix><command> [sub-command...] <usage-args>
//
// For example: !command delete <command>
func Format(cmd CommandStatic, prefix string) string {
	var args string
	if cmd.UsageArgs() != "" {
		args = " " + cmd.UsageArgs()
	}

	var path []string
	for cmd.Parent() != nil {
		path = append([]string{cmd.Names()[0]}, path...)
		cmd = cmd.Parent()
	}
	path = append([]string{cmd.Names()[0]}, path...)

	return prefix + strings.Join(path, " ") + args
}

// FormatQuote works much like Format but also quotes the resulting string using
// Messenger.QuoteCommand.
func FormatQuote(cmd CommandStatic, prefix string, msng Messenger) string {
	return msng.QuoteCommand(Format(cmd, prefix))
}

// CommandRuntime holds a command's runtime information.
type CommandRuntime struct {
	// The "path" taken to invoke the command, i.e., which names were used.
	// Includes all the sub-commands e.g. ["prefix", "add"] in order to be able
	// to display accurate help messages.
	Path []string

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
	usage := cmd.Prefix + strings.Join(cmd.Path, " ")
	if cmd.UsageArgs() != "" {
		usage += " " + cmd.UsageArgs()
	}
	return usage
}

type CommandsStatic []CommandStatic

func (cmds CommandsStatic) match(t CommandType, m *EventMessage, name string) (CommandStatic, error) {
	name = strings.ToLower(name)

	for _, c := range cmds {
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

// Match will return the corresponding command based on the list of arguments.
// The arguments don't have to match a command exactly, for example,
//
//	args = [prefix add abc]
//
// In this case, the prefix's subcommand "add" will be matched and returned.
// Alongside it, the index of the last valid command will be returned (in this
// case, the index of "add," which is 1).
func (cmds CommandsStatic) Match(t CommandType, m *EventMessage, args []string) (CommandStatic, int, error) {
	log.Debug().Strs("args", args).Msg("trying to match command")

	index := 0

	cmd, err := cmds.match(t, m, args[0])
	if err != nil {
		return nil, -1, err
	}

	for _, name := range args[1:] {
		tmp, err := cmd.Children().match(t, m, name)
		if err != nil {
			break
		}
		index++
		cmd = tmp
	}

	if !cmd.Permitted(m) {
		return nil, 0, errors.New("not permitted to execute command")
	}
	return cmd, index, nil
}

// Usage returns the names of the children in a format that can be used in the
// UsageArgs function.
func (cmds CommandsStatic) Usage() string {
	var names []string
	for _, c := range cmds {
		names = append(names, c.Names()[0])
	}
	return "(" + strings.Join(names, " | ") + ")"
}

// UsageOptional returns the names of the children in a format that can be used
// in the UsageArgs function and indicates that the sub-commands are optional.
func (cmds CommandsStatic) UsageOptional() string {
	var names []string
	for _, c := range cmds {
		names = append(names, c.Names()[0])
	}
	return "[" + strings.Join(names, " | ") + "]"
}

// Recurse will recursively go through all the commands and execute the exec
// function on them.
func (cmds CommandsStatic) Recurse(exec func(CommandStatic)) {
	for _, cmd := range cmds {
		exec(cmd)
		if cmd.Children() != nil {
			cmd.Children().Recurse(exec)
		}
	}
}
