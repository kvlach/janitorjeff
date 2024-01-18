package custom_command

import (
	"strings"

	"github.com/kvlach/janitorjeff/core"
)

var (
	UrrTriggerExists   = core.UrrNew("trigger already exists")
	UrrBuiltinCommand  = core.UrrNew("trigger collides with a built-in command")
	UrrTriggerNotFound = core.UrrNew("trigger was not found")
)

// Check if a string corresponds to a command name. Doesn't check sub-commands.
func isCommand(t core.CommandType, s string) bool {
	for _, c := range core.Commands {
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

func Add(place, creator int64, trigger, response string) (core.Urr, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if exists {
		return UrrTriggerExists, nil
	}

	builtin, err := isBuiltin(place, trigger)
	if err != nil {
		return nil, err
	}
	if builtin {
		return UrrBuiltinCommand, nil
	}

	return nil, dbAdd(place, creator, trigger, response)
}

func Edit(place, editor int64, trigger, response string) (core.Urr, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if !exists {
		return UrrTriggerNotFound, nil
	}
	return nil, dbEdit(place, editor, trigger, response)
}

func Delete(place, deleter int64, trigger string) (core.Urr, error) {
	exists, err := dbTriggerExists(place, trigger)
	if err != nil {
		return nil, err
	}
	if !exists {
		return UrrTriggerNotFound, nil
	}
	return nil, dbDelete(place, deleter, trigger)
}

func List(place int64) ([]string, error) {
	return dbList(place)
}

func Show(place int64, trigger string) (string, error) {
	return dbGetResponse(place, trigger)
}

func History(place int64, trigger string) ([]customCommand, error) {
	// We don't check to see if the trigger exists since this command may be
	// used to view the history of a deleted trigger
	return dbHistory(place, trigger)
}
