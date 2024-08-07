package prefix

import (
	"github.com/kvlach/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.EventMessage) bool {
	mod, err := m.Author.Moderator()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if author is mod")
		return false
	}
	return mod
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

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryModerators
}

func (advanced) Examples() []string {
	return nil
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
	core.EventMessageHooks.Register(c.emergencyReset)
	return nil
}

func (advanced) emergencyReset(m *core.EventMessage) {
	if m.Raw != "!!!PleaseResetThePrefixesBackToTheDefaultsThanks!!!" {
		return
	}

	resp, urr, err := AdvancedReset.Run(m)
	if err != nil {
		return
	}

	m.Write(resp, urr)
}

func (advanced) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c advancedAdd) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedAdd) Names() []string {
	return core.AliasesAdd
}

func (advancedAdd) Description() string {
	return "Add a prefix."
}

func (advancedAdd) UsageArgs() string {
	return "<prefix>"
}

func (c advancedAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedAdd) Examples() []string {
	return nil
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

func (c advancedAdd) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedDelete) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedDelete) Names() []string {
	return core.AliasesDelete
}

func (advancedDelete) Description() string {
	return "Delete a prefix."
}

func (advancedDelete) UsageArgs() string {
	return "<prefix>"
}

func (c advancedDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedDelete) Examples() []string {
	return nil
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

func (c advancedDelete) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedList) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedList) Names() []string {
	return core.AliasesList
}

func (advancedList) Description() string {
	return "List the current prefixes."
}

func (advancedList) UsageArgs() string {
	return ""
}

func (c advancedList) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedList) Examples() []string {
	return nil
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

func (c advancedList) Run(m *core.EventMessage) (any, core.Urr, error) {
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

func (c advancedReset) Permitted(m *core.EventMessage) bool {
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

func (c advancedReset) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedReset) Examples() []string {
	return nil
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

func (advancedReset) Run(m *core.EventMessage) (any, core.Urr, error) {
	return cmdReset(m)
}
