package prefix

import (
	"github.com/janitorjeff/jeff-bot/core"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return m.Author.Mod()
}

func (advanced) Names() []string {
	return []string{
		"prefix",
	}
}

func (advanced) Description() string {
	return "Add, delete, list or reset prefixes."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedAdd,
		AdvancedDelete,
		AdvancedList,
		AdvancedReset,
	}
}

func (c advanced) Init() error {
	core.Hooks.Register(c.emergencyReset)
	return nil
}

func (advanced) emergencyReset(m *core.Message) {
	if m.Raw != "!!!PleaseResetThePrefixesBackToTheDefaultsThanks!!!" {
		return
	}

	resp, usrErr, err := AdvancedReset.Run(m)
	if err != nil {
		return
	}

	m.Write(resp, usrErr)
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////
//     //
// add //
//     //
/////////

var AdvancedAdd = advancedAdd{}

type advancedAdd struct{}

func (c advancedAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAdd) Names() []string {
	return core.Add
}

func (advancedAdd) Description() string {
	return "Add a prefix."
}

func (advancedAdd) UsageArgs() string {
	return "<prefix>"
}

func (advancedAdd) Parent() core.CommandStatic {
	return Advanced
}

func (advancedAdd) Children() core.CommandsStatic {
	return nil
}

func (advancedAdd) Init() error {
	return nil
}

func (c advancedAdd) Run(m *core.Message) (any, error, error) {
	return cmdAdd(c.Type(), m)
}

////////////
//        //
// delete //
//        //
////////////

var AdvancedDelete = advancedDelete{}

type advancedDelete struct{}

func (c advancedDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDelete) Names() []string {
	return core.Delete
}

func (advancedDelete) Description() string {
	return "Delete a prefix."
}

func (advancedDelete) UsageArgs() string {
	return "<prefix>"
}

func (advancedDelete) Parent() core.CommandStatic {
	return Advanced
}

func (advancedDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedDelete) Init() error {
	return nil
}

func (c advancedDelete) Run(m *core.Message) (any, error, error) {
	return cmdDelete(c.Type(), m)
}

//////////
//      //
// list //
//      //
//////////

var AdvancedList = advancedList{}

type advancedList struct{}

func (c advancedList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedList) Names() []string {
	return core.List
}

func (advancedList) Description() string {
	return "List the current prefixes."
}

func (advancedList) UsageArgs() string {
	return ""
}

func (advancedList) Parent() core.CommandStatic {
	return Advanced
}

func (advancedList) Children() core.CommandsStatic {
	return nil
}

func (advancedList) Init() error {
	return nil
}

func (c advancedList) Run(m *core.Message) (any, error, error) {
	return cmdList(c.Type(), m)
}

///////////
//       //
// reset //
//       //
///////////

var AdvancedReset = advancedReset{}

type advancedReset struct{}

func (c advancedReset) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReset) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReset) Names() []string {
	return []string{
		"reset",
	}
}

func (advancedReset) Description() string {
	return "Reset prefixes to bot defaults."
}

func (advancedReset) UsageArgs() string {
	return ""
}

func (advancedReset) Parent() core.CommandStatic {
	return Advanced
}

func (advancedReset) Children() core.CommandsStatic {
	return nil
}

func (advancedReset) Init() error {
	return nil
}

func (advancedReset) Run(m *core.Message) (any, error, error) {
	return cmdReset(m)
}
