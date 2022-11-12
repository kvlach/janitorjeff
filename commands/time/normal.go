package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return Advanced.Frontends()
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return "Time stuff and things."
}

func (normal) UsageArgs() string {
	return "[<user> | (zone)]"
}

func (normal) Parent() core.Commander {
	return nil
}

func (normal) Children() core.Commanders {
	return core.Commanders{
		NormalZone,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedNow.Run(m)
}

//////////
//      //
// zone //
//      //
//////////

var NormalZone = normalZone{}

type normalZone struct{}

func (c normalZone) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalZone) Frontends() int {
	return c.Parent().Frontends()
}

func (normalZone) Names() []string {
	return []string{
		"zone",
	}
}

func (normalZone) Description() string {
	return "Set or view your own timezone."
}

func (normalZone) UsageArgs() string {
	return "[timezone]"
}

func (normalZone) Parent() core.Commander {
	return Normal
}

func (normalZone) Children() core.Commanders {
	return nil
}

func (normalZone) Init() error {
	return nil
}

func (normalZone) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedTimezoneShow.Run(m)
	}
	return AdvancedTimezoneSet.Run(m)
}
