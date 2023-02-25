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
