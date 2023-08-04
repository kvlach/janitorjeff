package god

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
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

func (normal) Examples() []string {
	return nil
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalShow,
		NormalOn,
		NormalOff,
		NormalInterval,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedTalk.Run(m)
}

//////////
//      //
// show //
//      //
//////////

var NormalShow = normalShow{}

type normalShow struct{}

func (c normalShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalShow) Names() []string {
	return AdvancedReplyShow.Names()
}

func (normalShow) Description() string {
	return AdvancedReplyShow.Description()
}

func (c normalShow) UsageArgs() string {
	return AdvancedReplyShow.UsageArgs()
}

func (c normalShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalShow) Examples() []string {
	return nil
}

func (normalShow) Parent() core.CommandStatic {
	return Normal
}

func (normalShow) Children() core.CommandsStatic {
	return nil
}

func (normalShow) Init() error {
	return nil
}

func (normalShow) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedReplyShow.Run(m)
}

////////
//    //
// on //
//    //
////////

var NormalOn = normalOn{}

type normalOn struct{}

func (c normalOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalOn) Names() []string {
	return AdvancedReplyOn.Names()
}

func (normalOn) Description() string {
	return AdvancedReplyOn.Description()
}

func (c normalOn) UsageArgs() string {
	return AdvancedReplyOn.UsageArgs()
}

func (c normalOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalOn) Examples() []string {
	return nil
}

func (normalOn) Parent() core.CommandStatic {
	return Normal
}

func (normalOn) Children() core.CommandsStatic {
	return nil
}

func (normalOn) Init() error {
	return nil
}

func (normalOn) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedReplyOn.Run(m)
}

/////////
//     //
// off //
//     //
/////////

var NormalOff = normalOff{}

type normalOff struct{}

func (c normalOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalOff) Names() []string {
	return AdvancedReplyOff.Names()
}

func (normalOff) Description() string {
	return AdvancedReplyOff.Description()
}

func (c normalOff) UsageArgs() string {
	return AdvancedReplyOff.UsageArgs()
}

func (c normalOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalOff) Examples() []string {
	return nil
}

func (normalOff) Parent() core.CommandStatic {
	return Normal
}

func (normalOff) Children() core.CommandsStatic {
	return nil
}

func (normalOff) Init() error {
	return nil
}

func (normalOff) Run(m *core.Message) (any, core.Urr, error) {
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

func (normalInterval) Examples() []string {
	return nil
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

func (normalInterval) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedIntervalShow.Run(m)
	}
	return AdvancedIntervalSet.Run(m)
}
