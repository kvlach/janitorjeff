package discord

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
)

// Here implements the core.Here interface.
type Here struct {
	ChannelID string
	GuildID   string
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
	rdbKey := "frontend_discord_scope_here_exact_" + h.IDExact()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceExactScope(h.IDExact(), h.ChannelID, h.GuildID)
	})
}

func (h *Here) ScopeLogical() (int64, error) {
	rdbKey := "frontend_discord_scope_here_logical_" + h.IDExact()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceLogicalScope(h.IDExact(), h.ChannelID, h.GuildID)
	})
}
