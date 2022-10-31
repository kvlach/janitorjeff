package urban_dictionary

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"ud",
	},
	Description: "Search a term on urban dictionary.",
	UsageArgs:   "<term...>",
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	return advancedRunSearch(m)
}
