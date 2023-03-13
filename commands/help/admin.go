package help

import (
	"github.com/janitorjeff/jeff-bot/core"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.Message) bool {
	return true
}

func (admin) Names() []string {
	return cmdNames
}

func (admin) Description() string {
	return "Shows a help message for the specified admin command."
}

func (admin) UsageArgs() string {
	return cmdUsageArgs
}

func (admin) Category() core.CommandCategory {
	return Advanced.Category()
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return nil
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return run(core.Admin, m)
}
