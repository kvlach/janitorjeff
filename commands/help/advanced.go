package help

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.Type {
	return core.Advanced
}

func (advanced) Frontends() int {
	return frontends.All
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

func (advanced) Parent() core.Commander {
	return nil
}

func (advanced) Children() core.Commanders {
	return nil
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return run(core.Advanced, m)
}
