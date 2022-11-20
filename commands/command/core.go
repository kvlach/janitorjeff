package command

import (
	"errors"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var (
	errTriggerExists   = errors.New("trigger already exists")
	errBuiltinCommand  = errors.New("trigger collides with a built-in command")
	errTriggerNotFound = errors.New("trigger was not found")
)

// Check if a string corresponds to a command name. Doesn't check sub-commands.
func isCommand(t core.CommandType, s string) bool {
	for _, c := range *core.Commands {
		for _, n := range c.Names() {
			if c.Type() == t && n == s {
				return true
			}
		}
	}
	return false
}

func isBuiltin(place int64, trigger string) (bool, error) {
	prefixes, _, err := core.PlacePrefixes(place)
	if err != nil {
		return false, err
	}

	for _, p := range prefixes {
		// only check if there is a collision if the trigger begins with a
		// prefix used by the builtin commands
		if !strings.HasPrefix(trigger, p.Prefix) {
			continue
		}

		if isCommand(p.Type, strings.TrimPrefix(trigger, p.Prefix)) {
			return true, nil
		}
	}

	return false, nil
}

func Add(place, creator int64, trigger, response string) (error, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if exists {
		return errTriggerExists, nil
	}

	builtin, err := isBuiltin(place, trigger)
	if err != nil {
		return nil, err
	}
	if builtin {
		return errBuiltinCommand, nil
	}

	return nil, dbAdd(place, creator, trigger, response)
}

func Edit(place, editor int64, trigger, response string) (error, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errTriggerNotFound, nil
	}
	return nil, dbEdit(place, editor, trigger, response)
}

func Delete(place, deleter int64, trigger string) (error, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errTriggerNotFound, nil
	}
	return nil, dbDelete(place, deleter, trigger)
}

func List(place int64) ([]string, error) {
	return dbList(place)
}

func History(place int64, trigger string) ([]customCommand, error) {
	// We don't check to see if the trigger exists since this command may be
	// used to view the history of a deleted trigger
	return dbHistory(place, trigger)
}
