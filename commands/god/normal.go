package god

import (
	"github.com/janitorjeff/jeff-bot/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.Message) bool {
	return Advanced.Permitted(m)
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return Advanced.Description()
}

func (c normal) UsageArgs() string {
	return c.Children().UsageOptional()
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalInterval,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedTalk.Run(m)
}

//////////////
//          //
// interval //
//          //
//////////////

var NormalInterval = normalInterval{}

type normalInterval struct{}

func (c normalInterval) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalInterval) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalInterval) Names() []string {
	return AdvancedInterval.Names()
}

func (normalInterval) Description() string {
	return AdvancedInterval.Description()
}

func (normalInterval) UsageArgs() string {
	return "[interval]"
}

func (normalInterval) Parent() core.CommandStatic {
	return Normal
}

func (normalInterval) Children() core.CommandsStatic {
	return nil
}

func (normalInterval) Init() error {
	return nil
}

func (normalInterval) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedIntervalShow.Run(m)
	}
	return AdvancedIntervalSet.Run(m)
}
