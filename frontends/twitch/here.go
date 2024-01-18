package twitch

import (
	"errors"

	"github.com/kvlach/janitorjeff/core"
)

// here implements the core.Placer interface.
type here struct {
	roomID   string
	roomName string
	author   core.Personifier

	scopeCache int64
}

// NewHere initializes a here object.
// If roomID is unknown, pass an empty string.
// If roomName is unknown, pass an empty string.
// At least one of roomID or roomName must not be an empty string.
// The author must not be nil.
func NewHere(roomID, roomName string, author core.Personifier) (core.Placer, error) {
	if roomID == "" && roomName == "" {
		return nil, errors.New("at least one of roomID or roomName is required")
	}
	if author == nil {
		return nil, errors.New("author must not be nil")
	}
	return here{
		roomID:   roomID,
		roomName: roomName,
		author:   author,
	}, nil
}

func (h here) IDExact() (string, error) {
	if h.roomID != "" {
		return h.roomID, nil
	}

	hx, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	uid, err := hx.GetUserID(h.roomName)
	if err != nil {
		return "", err
	}
	h.roomID = uid
	return uid, nil
}

func (h here) IDLogical() (string, error) {
	return h.IDExact()
}

func (h here) Name() (string, error) {
	if h.roomName != "" {
		return h.roomName, nil
	}

	hx, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	u, err := hx.GetUser(h.roomID)
	if err != nil {
		return "", err
	}
	h.roomName = u.Login
	return u.Login, nil
}

func (h here) scope() (int64, error) {
	if h.scopeCache != 0 {
		return h.scopeCache, nil
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
	h.scopeCache = scope
	return scope, nil
}

func (h here) ScopeExact() (int64, error) {
	author, err := h.author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Exact, nil
	}
	return h.scope()
}

func (h here) ScopeLogical() (int64, error) {
	author, err := h.author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Logical, nil
	}
	return h.scope()
}
