package core

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type Commands []*CommandStatic

func (cmds *Commands) matchCommand(cmdName string) (*CommandStatic, error) {
	if cmdName == "" {
		return nil, fmt.Errorf("no command provided")
	}

	cmdName = strings.ToLower(cmdName)

	for _, c := range *cmds {
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
func (cmds *Commands) MatchCommand(args []string) (*CommandStatic, int, error) {
	log.Debug().Strs("args", args).Msg("trying to match command")

	index := 0

	cmd, err := cmds.matchCommand(args[index])
	if err != nil {
		return nil, -1, err
	}

	for _, c := range args[1:] {
		tmp, err := cmd.Children.matchCommand(c)
		if err != nil {
			return cmd, index, nil
		}
		index++
		cmd = tmp
	}

	return cmd, index, nil
}
