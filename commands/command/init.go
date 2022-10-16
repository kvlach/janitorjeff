package command

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func init_() error {
	core.Globals.Hooks.Register(writeCustomCommand)
	return core.Globals.DB.Init(dbShema)
}

func writeCustomCommand(m *core.Message) {
	fields := m.Fields()

	if len(fields) > 1 {
		return
	}

	scope, err := m.ScopePlace()
	if err != nil {
		return
	}

	resp, err := dbGetResponse(scope, fields[0])
	if err != nil {
		return
	}

	m.Write(resp, nil)
}
