package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func init_() error {
	return core.Globals.DB.Init(dbSchema)
}
