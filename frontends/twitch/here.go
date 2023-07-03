package twitch

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
)

type Here struct {
	RoomID   string
	RoomName string
}

func (h Here) IDExact() string {
	return h.RoomID
}

func (h Here) IDLogical() string {
	return h.RoomID
}

func (h Here) Name() string {
	return h.RoomName
}

func (h Here) Scope() (int64, error) {
	rdbKey := "frontend_twitch_scope_" + h.IDExact()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbAddChannelSimple(h.IDExact(), h.Name())
	})
}

func (h Here) ScopeExact() (int64, error) {
	return h.Scope()
}

func (h Here) ScopeLogical() (int64, error) {
	return h.Scope()
}
