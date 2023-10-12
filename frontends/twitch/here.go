package twitch

import (
	"errors"

	"git.sr.ht/~slowtyper/janitorjeff/core"
)

type Here struct {
	RoomID   string
	RoomName string
	Author   core.Personifier

	scope int64
}

func (h Here) IDExact() (string, error) {
	return h.RoomID, nil
}

func (h Here) IDLogical() (string, error) {
	return h.RoomID, nil
}

func (h Here) Name() (string, error) {
	if h.RoomName != "" {
		return h.RoomName, nil
	}

	if h.RoomID == "" {
		return "", errors.New("the room id is required")
	}
	hx, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	u, err := hx.GetUser(h.RoomID)
	if err != nil {
		return "", err
	}
	return u.Login, nil
}

func (h Here) Scope() (int64, error) {
	if h.scope != 0 {
		return h.scope, nil
	}
	hix, err := h.IDExact()
	if err != nil {
		return 0, err
	}
	rdbKey := "frontend_twitch_scope_" + hix
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return dbAddChannel(hix)
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
