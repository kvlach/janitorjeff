package streak

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
		NormalOn,
		NormalOff,
		NormalRedeem,
		NormalGrace,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedShow.Run(m)
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
	return AdvancedOn.Names()
}

func (normalOn) Description() string {
	return AdvancedOn.Description()
}

func (normalOn) UsageArgs() string {
	return AdvancedOn.UsageArgs()
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

func (normalOn) Run(m *core.Message) (any, error, error) {
	return AdvancedOn.Run(m)
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
	return AdvancedOff.Names()
}

func (normalOff) Description() string {
	return AdvancedOff.Description()
}

func (normalOff) UsageArgs() string {
	return AdvancedOff.UsageArgs()
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

func (normalOff) Run(m *core.Message) (any, error, error) {
	return AdvancedOff.Run(m)
}

////////////
//        //
// redeem //
//        //
////////////

var NormalRedeem = normalRedeem{}

type normalRedeem struct{}

func (c normalRedeem) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalRedeem) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalRedeem) Names() []string {
	return AdvancedRedeem.Names()
}

func (normalRedeem) Description() string {
	return AdvancedRedeem.Description()
}

func (normalRedeem) UsageArgs() string {
	return "[id]"
}

func (c normalRedeem) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalRedeem) Examples() []string {
	return nil
}

func (normalRedeem) Parent() core.CommandStatic {
	return Normal
}

func (normalRedeem) Children() core.CommandsStatic {
	return nil
}

func (normalRedeem) Init() error {
	return nil
}

func (normalRedeem) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedRedeemShow.Run(m)
	}
	return AdvancedRedeemSet.Run(m)
}

///////////
//       //
// grace //
//       //
///////////

var NormalGrace = normalGrace{}

type normalGrace struct{}

func (c normalGrace) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalGrace) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalGrace) Names() []string {
	return AdvancedGrace.Names()
}

func (normalGrace) Description() string {
	return AdvancedGrace.Description()
}

func (normalGrace) UsageArgs() string {
	return "[duration]"
}

func (c normalGrace) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalGrace) Examples() []string {
	return nil
}

func (normalGrace) Parent() core.CommandStatic {
	return Normal
}

func (normalGrace) Children() core.CommandsStatic {
	return nil
}

func (normalGrace) Init() error {
	return nil
}

func (normalGrace) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedGraceShow.Run(m)
	}
	return AdvancedGraceSet.Run(m)
}
