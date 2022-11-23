package prefix

import (
	"errors"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/command"
	"github.com/janitorjeff/jeff-bot/core"
)

var (
	errExists   = errors.New("The prefix already exists.")
	errNotFound = errors.New("The prefix was not found.")
	errOneLeft  = errors.New("Only one prefix is left.")

	errCustomCommandExists = errors.New("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.")
)

func isCommand(s string) bool {
	for _, c := range *core.Commands {
		for _, n := range c.Names() {
			if n == s {
				return true
			}
		}
	}
	return false
}

// if the prefix changes after a custom command has been added it's
// possible that a collision may be created
//
// for example:
// !prefix reset
// !cmd add .prefix test // this works because . is not a valid prefix atm
// !prefix add .
// .prefix // both trigger
func customCommandCollision(prefix string, place int64) (string, error) {
	triggers, err := command.List(place)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		if isCommand(strings.TrimPrefix(t, prefix)) {
			return prefix + t, nil
		}
	}

	return "", nil
}

func Add(prefix string, t core.CommandType, place int64) (string, error, error) {
	prefixes, inDB, err := core.PlacePrefixes(place)
	if err != nil {
		return "", nil, err
	}

	// Only add default prefixes if they've never been added before, this
	// prevents situations were the default prefixes change and they sneakily
	// get added without the user realizing.
	if !inDB {
		for _, p := range prefixes {
			if err = dbAdd(p.Prefix, p.Type, place); err != nil {
				return "", nil, err
			}
		}
	}

	for _, p := range prefixes {
		if p.Prefix == prefix {
			return "", errExists, nil
		}
	}

	collision, err := customCommandCollision(prefix, place)
	if err != nil {
		return "", nil, err
	}
	if collision != "" {
		return collision, errCustomCommandExists, nil
	}

	return "", nil, dbAdd(prefix, t, place)
}

func Delete(prefix string, t core.CommandType, place int64) (error, error) {
	prefixes, inDB, err := core.PlacePrefixes(place)
	if err != nil {
		return nil, err
	}

	exists := false
	for _, p := range prefixes {
		if t&p.Type == 0 {
			continue
		}

		if p.Prefix == prefix {
			exists = true
		}
	}

	if !exists {
		return errNotFound, nil
	}
	if len(prefixes) == 1 {
		return errOneLeft, nil
	}

	// If the scope doesn't exist then the default prefixes are being used and
	// they are not present in the DB. So if the user tries to delete one
	// nothing will happen. So we first add them all to the DB.
	if !inDB {
		for _, p := range prefixes {
			if err = dbAdd(p.Prefix, p.Type, place); err != nil {
				return nil, err
			}
		}
	}

	return nil, dbDelete(prefix, place)
}

func List(t core.CommandType, place int64) ([]core.Prefix, error) {
	prefixes, _, err := core.PlacePrefixes(place)

	ps := []core.Prefix{}
	for _, p := range prefixes {
		if t&p.Type != 0 {
			ps = append(ps, p)
		}
	}

	return ps, err
}

func Reset(place int64) error {
	return dbReset(place)
}
