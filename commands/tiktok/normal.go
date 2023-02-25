package tiktok

import (
	"github.com/janitorjeff/jeff-bot/core"
)

///////////
//       //
// voice //
//       //
///////////

// NormalVoice is not a sub-command
var NormalVoice = normalVoice{}

type normalVoice struct{}

func (normalVoice) Type() core.CommandType {
	return core.Normal
}

func (normalVoice) Permitted(m *core.Message) bool {
	return AdvancedVoice.Permitted(m)
}

func (normalVoice) Names() []string {
	return AdvancedVoice.Names()
}

func (normalVoice) Description() string {
	return AdvancedVoice.Description()
}

func (normalVoice) UsageArgs() string {
	return "<user> [voice]"
}

func (normalVoice) Parent() core.CommandStatic {
	return nil
}

func (normalVoice) Children() core.CommandsStatic {
	return nil
}

func (normalVoice) Init() error {
	return nil
}

func (normalVoice) Run(m *core.Message) (any, error, error) {
	switch len(m.Command.Args) {
	case 0:
		return m.Usage(), core.ErrMissingArgs, nil
	case 1:
		return AdvancedVoiceShow.Run(m)
	default:
		return AdvancedVoiceSet.Run(m)
	}
}

/////////////
//         //
// subonly //
//         //
/////////////

var NormalSubOnly = normalSubOnly{}

type normalSubOnly struct{}

func (normalSubOnly) Type() core.CommandType {
	return core.Normal
}

func (normalSubOnly) Permitted(m *core.Message) bool {
	return AdvancedSubOnly.Permitted(m)
}

func (normalSubOnly) Names() []string {
	return AdvancedSubOnly.Names()
}

func (normalSubOnly) Description() string {
	return AdvancedSubOnly.Description()
}

func (c normalSubOnly) UsageArgs() string {
	return c.Children().UsageOptional()
}

func (normalSubOnly) Parent() core.CommandStatic {
	return nil
}

func (normalSubOnly) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalSubOnlyOn,
		NormalSubOnlyOff,
	}
}

func (normalSubOnly) Init() error {
	return nil
}

func (normalSubOnly) Run(m *core.Message) (any, error, error) {
	return AdvancedSubOnlyShow.Run(m)
}

////////////////
//            //
// subonly on //
//            //
////////////////

var NormalSubOnlyOn = normalSubOnlyOn{}

type normalSubOnlyOn struct{}

func (normalSubOnlyOn) Type() core.CommandType {
	return core.Normal
}

func (normalSubOnlyOn) Permitted(m *core.Message) bool {
	return AdvancedSubOnlyOn.Permitted(m)
}

func (normalSubOnlyOn) Names() []string {
	return AdvancedSubOnlyOn.Names()
}

func (normalSubOnlyOn) Description() string {
	return AdvancedSubOnlyOn.Description()
}

func (normalSubOnlyOn) UsageArgs() string {
	return ""
}

func (normalSubOnlyOn) Parent() core.CommandStatic {
	return NormalSubOnly
}

func (normalSubOnlyOn) Children() core.CommandsStatic {
	return nil
}

func (normalSubOnlyOn) Init() error {
	return nil
}

func (normalSubOnlyOn) Run(m *core.Message) (any, error, error) {
	return AdvancedSubOnlyOn.Run(m)
}

/////////////////
//             //
// subonly off //
//             //
/////////////////

var NormalSubOnlyOff = normalSubOnlyOff{}

type normalSubOnlyOff struct{}

func (normalSubOnlyOff) Type() core.CommandType {
	return core.Normal
}

func (normalSubOnlyOff) Permitted(m *core.Message) bool {
	return AdvancedSubOnlyOff.Permitted(m)
}

func (normalSubOnlyOff) Names() []string {
	return AdvancedSubOnlyOff.Names()
}

func (normalSubOnlyOff) Description() string {
	return AdvancedSubOnlyOff.Description()
}

func (normalSubOnlyOff) UsageArgs() string {
	return ""
}

func (normalSubOnlyOff) Parent() core.CommandStatic {
	return NormalSubOnly
}

func (normalSubOnlyOff) Children() core.CommandsStatic {
	return nil
}

func (normalSubOnlyOff) Init() error {
	return nil
}

func (normalSubOnlyOff) Run(m *core.Message) (any, error, error) {
	return AdvancedSubOnlyOff.Run(m)
}
