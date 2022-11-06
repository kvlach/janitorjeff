package urban_dictionary

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"ud",
	},
	Description: "Search a term on urban dictionary.",
	UsageArgs:   "<term...>",
	Frontends:   frontends.All,
	Run:         normalRun,
}

func normalRun(m *core.Message) (any, error, error) {
	return advancedRunSearch(m)
}
