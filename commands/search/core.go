package search

import (
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
)

// Recurse will recursively go through all of the commands and run the given
// check function on them. If the returned value is true then that specific
// command is added to the returned list. If both a parent and their child are
// matched then only the child is added to the list.
func Recurse(check func(core.CommandStatic) bool) core.CommandsStatic {
	var matched core.CommandsStatic

	core.Commands.Recurse(func(cmd core.CommandStatic) {
		if !check(cmd) {
			return
		}

		if len(matched) == 0 {
			matched = append(matched, cmd)
			return
		}

		// Common usage args pattern for parents is to include their
		// children's names, so this results in both the parent and the
		// child being matched which just bloats the list of returned
		// commands. We fix that by replacing the parent with the child
		// instead of blindly appending the child.
		if matched[len(matched)-1] == cmd.Parent() {
			matched[len(matched)-1] = cmd
		} else {
			matched = append(matched, cmd)
		}
	})

	return matched
}

// Search will recursively go through all of the commands and try to match them
// using the given query. It will check a command's names, description and
// usage args.
func Search(query string, t core.CommandType) core.CommandsStatic {
	tokens := strings.Fields(strings.ToLower(query))

	return Recurse(func(cmd core.CommandStatic) bool {
		if cmd.Type()&t == 0 {
			return false
		}

		desc := strings.ToLower(cmd.Description())

		for _, tk := range tokens {
			for _, n := range cmd.Names() {
				if tk == n {
					return true
				}
			}

			if strings.Contains(desc, tk) {
				return true
			}

			if strings.Contains(cmd.UsageArgs(), tk) {
				return true
			}
		}

		return false
	})
}
