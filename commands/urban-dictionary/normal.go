package urban_dictionary

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return AdvancedSearch.Frontends()
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return "Search a term on urban dictionary."
}

func (normal) UsageArgs() string {
	return "<term...>"
}

func (normal) Parent() core.Commander {
	return nil
}

func (normal) Children() core.Commanders {
	return nil
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedSearch.Run(m)
}
