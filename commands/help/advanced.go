package help

import (
	"github.com/janitorjeff/jeff-bot/core"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.Message) bool {
	return true
}

func (advanced) Names() []string {
	return cmdNames
}

func (advanced) Description() string {
	return "Shows a help message for the speicifed advanced command."
}

func (advanced) UsageArgs() string {
	return cmdUsageArgs
}

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return nil
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return run(core.Advanced, m)
}
