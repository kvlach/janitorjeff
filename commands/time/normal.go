package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(*core.Message) bool {
	return true
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

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
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

func (c normalZone) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
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

func (normalZone) Parent() core.CommandStatic {
	return Normal
}

func (normalZone) Children() core.CommandsStatic {
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
