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

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalReply,
		NormalInterval,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedTalk.Run(m)
}

///////////
//       //
// reply //
//       //
///////////

var NormalReply = normalReply{}

type normalReply struct{}

func (c normalReply) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalReply) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalReply) Names() []string {
	return AdvancedReply.Names()
}

func (normalReply) Description() string {
	return AdvancedReply.Description()
}

func (c normalReply) UsageArgs() string {
	return c.Children().UsageOptional()
}

func (c normalReply) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalReply) Parent() core.CommandStatic {
	return Normal
}

func (normalReply) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalReplyOn,
		NormalReplyOff,
	}
}

func (normalReply) Init() error {
	return nil
}

func (normalReply) Run(m *core.Message) (any, error, error) {
	return AdvancedReplyShow.Run(m)
}

//////////////
//          //
// reply on //
//          //
//////////////

var NormalReplyOn = normalReplyOn{}

type normalReplyOn struct{}

func (c normalReplyOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalReplyOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalReplyOn) Names() []string {
	return AdvancedReplyOn.Names()
}

func (normalReplyOn) Description() string {
	return AdvancedReplyOn.Description()
}

func (c normalReplyOn) UsageArgs() string {
	return AdvancedReplyOn.UsageArgs()
}

func (c normalReplyOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalReplyOn) Parent() core.CommandStatic {
	return NormalReply
}

func (normalReplyOn) Children() core.CommandsStatic {
	return nil
}

func (normalReplyOn) Init() error {
	return nil
}

func (normalReplyOn) Run(m *core.Message) (any, error, error) {
	return AdvancedReplyOn.Run(m)
}

///////////////
//           //
// reply off //
//           //
///////////////

var NormalReplyOff = normalReplyOff{}

type normalReplyOff struct{}

func (c normalReplyOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalReplyOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalReplyOff) Names() []string {
	return AdvancedReplyOff.Names()
}

func (normalReplyOff) Description() string {
	return AdvancedReplyOff.Description()
}

func (c normalReplyOff) UsageArgs() string {
	return AdvancedReplyOff.UsageArgs()
}

func (c normalReplyOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalReplyOff) Parent() core.CommandStatic {
	return NormalReply
}

func (normalReplyOff) Children() core.CommandsStatic {
	return nil
}

func (normalReplyOff) Init() error {
	return nil
}

func (normalReplyOff) Run(m *core.Message) (any, error, error) {
	return AdvancedReplyOff.Run(m)
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

func (c normalInterval) Category() core.CommandCategory {
	return c.Parent().Category()
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
