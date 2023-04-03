package discord

import (
	"github.com/janitorjeff/jeff-bot/core"
)

// Here implements the core.Here interface.
type Here struct {
	ChannelID string
	GuildID   string
}

func (h *Here) ID() string {
	return h.ChannelID
}

func (h *Here) Name() string {
	return h.ChannelID
}

func (h *Here) ScopeExact() (int64, error) {
	rdbKey := "frontend_discord_scope_here_exact_" + h.ID()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceExactScope(h.ID(), h.ChannelID, h.GuildID)
	})
}

func (h *Here) ScopeLogical() (int64, error) {
	rdbKey := "frontend_discord_scope_here_logical_" + h.ID()

	return core.CacheScope(rdbKey, func() (int64, error) {
		return getPlaceLogicalScope(h.ID(), h.ChannelID, h.GuildID)
	})
}
