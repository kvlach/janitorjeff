package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return frontends.All
}

func (normal) Permitted(*core.Message) bool {
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

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return nil
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return run(core.Normal, m)
}
