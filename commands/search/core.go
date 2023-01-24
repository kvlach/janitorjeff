package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
)

type Match struct {
	command    core.CommandStatic
	occurences int
}

// Recurse will recursively go through all of the commands and run the given
// check function on them. If the returned value is true then that specific
// command is added to the returned list. If both a parent and their child are
// matched then only the child is added to the list.
func Recurse(match func(core.CommandStatic) int) []Match {
	var matched []Match

	core.Commands.Recurse(func(cmd core.CommandStatic) {
		if matches := match(cmd); matches != 0 {
			matched = append(matched, Match{cmd, matches})
		}
	})

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].occurences > matched[j].occurences
	})

	return matched
}

// Search will recursively go through all of the commands and try to match them
// using the given query. It will check a command's names, description and
// usage args.
func Search(query string, t core.CommandType) []Match {
	tokens := strings.Fields(strings.ToLower(core.Clean(query)))

	fmt.Println(tokens)

	return Recurse(func(cmd core.CommandStatic) (cnt int) {
		if cmd.Type()&t == 0 {
			return
		}

		desc := strings.ToLower(cmd.Description())

		for _, tk := range tokens {
			for _, n := range cmd.Names() {
				if tk == n {
					cnt++
				}
			}

			if strings.Contains(desc, tk) {
				cnt++
			}

			if strings.Contains(cmd.UsageArgs(), tk) {
				cnt++
			}
		}

		return
	})
}
