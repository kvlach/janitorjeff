package help

import (
	"github.com/kvlach/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(*core.EventMessage) bool {
	return true
}

func (normal) Names() []string {
	return cmdNames
}

func (normal) Description() string {
	return "Shows a help message for the specified command."
}

func (normal) UsageArgs() string {
	return cmdUsageArgs
}

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Examples() []string {
	return nil
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

func (normal) Run(m *core.EventMessage) (any, core.Urr, error) {
	return run(core.Normal, m)
}
