package category

import (
	"github.com/kvlach/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.EventMessage) bool {
	return Advanced.Permitted(m)
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return Advanced.Description()
}

func (normal) UsageArgs() string {
	return "[category]"
}

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Examples() []string {
	return append([]string{""}, AdvancedEdit.Examples()...)
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return nil
}

func (normal) Init() error {
	return nil
}

func (c normal) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedShow.Run(m)
	}
	return AdvancedEdit.Run(m)
}
