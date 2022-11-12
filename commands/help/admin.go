package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Frontends() int {
	return frontends.All
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

func (admin) Parent() core.Commander {
	return nil
}

func (admin) Children() core.Commanders {
	return nil
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return run(core.Admin, m)
}
