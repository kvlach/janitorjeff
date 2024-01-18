package discord

import (
	"sync"

	"github.com/kvlach/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

// Here implements the core.Placer interface.
type Here struct {
	ChannelID string
	GuildID   string
	Author    core.Personifier

	mu           sync.Mutex
	scopeExact   int64
	scopeLogical int64
}

func (h *Here) IDExact() (string, error) {
	return h.ChannelID, nil
}

func (h *Here) IDLogical() (string, error) {
	if h.GuildID == "" {
		return h.ChannelID, nil
	}
	return h.GuildID, nil
}

func (h *Here) Name() (string, error) {
	return h.ChannelID, nil
}

func (h *Here) ScopeExact() (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	author, err := h.Author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Exact, nil
	}

	if h.scopeExact != 0 {
		log.Debug().
			Int64("scope", h.scopeLogical).
			Msg("LOCAL: found cached scope")
		return h.scopeExact, nil
	}

	hix, err := h.IDExact()
	if err != nil {
		return 0, err
	}
	rdbKey := "frontend_discord_scope_here_exact_" + hix
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceExactScope(hix)
	})
	if err != nil {
		return 0, err
	}
	h.scopeExact = scope
	return scope, nil
}

func (h *Here) ScopeLogical() (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	author, err := h.Author.Scope()
	if err != nil {
		return 0, err
	}
	if place, ok := core.Teleports.Get(author); ok {
		return place.Logical, nil
	}

	if h.scopeLogical != 0 {
		log.Debug().
			Int64("scope", h.scopeLogical).
			Msg("LOCAL: found cached scope")
		return h.scopeLogical, nil
	}

	hix, err := h.IDExact()
	if err != nil {
		return 0, err
	}
	rdbKey := "frontend_discord_scope_here_logical_" + hix
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceLogicalScope(hix)
	})
	if err != nil {
		return 0, err
	}
	h.scopeLogical = scope
	return scope, nil
}
