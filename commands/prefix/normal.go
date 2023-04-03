package prefix

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
	return "[" + c.Children().Usage() + "]"
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
		NormalAdd,
		NormalDelete,
		NormalReset,
	}
}

func (normal) Init() error {
	return nil
}

func (c normal) Run(m *core.Message) (any, error, error) {
	return cmdList(c.Type(), m)
}

/////////
//     //
// add //
//     //
/////////

var NormalAdd = normalAdd{}

type normalAdd struct{}

func (c normalAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalAdd) Names() []string {
	return AdvancedAdd.Names()
}

func (normalAdd) Description() string {
	return AdvancedAdd.Description()
}

func (normalAdd) UsageArgs() string {
	return AdvancedAdd.UsageArgs()
}

func (c normalAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalAdd) Examples() []string {
	return nil
}

func (normalAdd) Parent() core.CommandStatic {
	return Normal
}

func (normalAdd) Children() core.CommandsStatic {
	return nil
}

func (normalAdd) Init() error {
	return nil
}

func (c normalAdd) Run(m *core.Message) (any, error, error) {
	return cmdAdd(c.Type(), m)
}

////////////
//        //
// delete //
//        //
////////////

var NormalDelete = normalDelete{}

type normalDelete struct{}

func (c normalDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalDelete) Names() []string {
	return AdvancedDelete.Names()
}

func (normalDelete) Description() string {
	return AdvancedDelete.Description()
}

func (normalDelete) UsageArgs() string {
	return AdvancedDelete.UsageArgs()
}

func (c normalDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalDelete) Examples() []string {
	return nil
}

func (normalDelete) Parent() core.CommandStatic {
	return Normal
}

func (normalDelete) Children() core.CommandsStatic {
	return nil
}

func (normalDelete) Init() error {
	return nil
}

func (c normalDelete) Run(m *core.Message) (any, error, error) {
	return cmdDelete(c.Type(), m)
}

///////////
//       //
// reset //
//       //
///////////

var NormalReset = normalReset{}

type normalReset struct{}

func (c normalReset) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalReset) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalReset) Names() []string {
	return AdvancedReset.Names()
}

func (normalReset) Description() string {
	return AdvancedReset.Description()
}

func (normalReset) UsageArgs() string {
	return AdvancedReset.UsageArgs()
}

func (c normalReset) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalReset) Examples() []string {
	return nil
}

func (normalReset) Parent() core.CommandStatic {
	return Normal
}

func (normalReset) Children() core.CommandsStatic {
	return nil
}

func (normalReset) Init() error {
	return nil
}

func (normalReset) Run(m *core.Message) (any, error, error) {
	return cmdReset(m)
}
