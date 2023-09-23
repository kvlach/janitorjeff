package twitch

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
)

type Here struct {
	RoomID   string
	RoomName string
	Author   core.Personifier

	scope int64
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
	if h.scope != 0 {
		return h.scope, nil
	}
	rdbKey := "frontend_twitch_scope_" + h.IDExact()
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return dbAddChannelSimple(h.IDExact(), h.Name())
	})
	if err != nil {
		return 0, err
	}
	h.scope = scope
	return scope, nil
}

func (h Here) ScopeExact() (int64, error) {
	author, err := h.Author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Exact, nil
	}
	return h.Scope()
}

func (h Here) ScopeLogical() (int64, error) {
	author, err := h.Author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Logical, nil
	}
	return h.Scope()
}
