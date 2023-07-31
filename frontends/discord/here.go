package discord

import (
	"sync"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

// Here implements the core.Placer interface.
type Here struct {
	ChannelID string
	GuildID   string

	mu           sync.Mutex
	scopeExact   int64
	scopeLogical int64
}

func (h *Here) IDExact() string {
	return h.ChannelID
}

func (h *Here) IDLogical() string {
	if h.GuildID == "" {
		return h.ChannelID
	}
	return h.GuildID
}

func (h *Here) Name() string {
	return h.ChannelID
}

func (h *Here) ScopeExact() (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.scopeExact != 0 {
		log.Debug().
			Int64("scope", h.scopeLogical).
			Msg("LOCAL: found cached scope")
		return h.scopeExact, nil
	}

	rdbKey := "frontend_discord_scope_here_exact_" + h.IDExact()
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceExactScope(h.IDExact(), h.ChannelID, h.GuildID)
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

	if h.scopeLogical != 0 {
		log.Debug().
			Int64("scope", h.scopeLogical).
			Msg("LOCAL: found cached scope")
		return h.scopeLogical, nil
	}

	rdbKey := "frontend_discord_scope_here_logical_" + h.IDExact()
	scope, err := core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceLogicalScope(h.IDExact(), h.ChannelID, h.GuildID)
	})
	if err != nil {
		return 0, err
	}
	h.scopeLogical = scope
	return scope, nil
}
