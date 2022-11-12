package nick

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

func (normal) Names() []string {
	return []string{
		"nick",
		"nickname",
	}
}

func (normal) Description() string {
	return "View or set your nickname."
}

func (normal) UsageArgs() string {
	return "[nickname]"
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
	if len(m.Command.Args) == 0 {
		return AdvancedShow.Run(m)
	}
	return AdvancedSet.Run(m)
}
