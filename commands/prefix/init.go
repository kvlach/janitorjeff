package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func init_() error {
	core.Globals.Hooks.Register(emergencyReset)
	return nil
}

func emergencyReset(m *core.Message) {
	if m.Raw != "!!!PleaseResetThePrefixesBackToTheDefaultsThanks!!!" {
		return
	}

	resp, usrErr, err := runReset(m)
	if err != nil {
		return
	}

	m.Write(resp, usrErr)
}
