package twitch

import (
	"github.com/janitorjeff/jeff-bot/core"
)

type Here struct {
	RoomID   string
	RoomName string
}

func (h Here) ID() string {
	return h.RoomID
}

func (h Here) Name() string {
	return h.RoomName
}

func (h Here) Scope() (int64, error) {
	rdbKey := "frontend_twitch_scope_" + h.ID()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbAddChannelSimple(h.ID(), h.Name())
	})
}

func (h Here) ScopeExact() (int64, error) {
	return h.Scope()
}

func (h Here) ScopeLogical() (int64, error) {
	return h.Scope()
}
