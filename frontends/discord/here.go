package discord

import (
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
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
	slog := log.With().Str("id", h.ID()).Logger()
	rdbKey := "frontend_discord_scope_here_exact_" + h.ID()

	scope, err := core.RDB.Get(ctx, rdbKey).Int64()
	if err != nil && err != redis.Nil {
		return -1, err
	}
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = getPlaceExactScope(h.ID(), h.ChannelID, h.GuildID)
	if err != nil {
		return -1, err
	}
	err = core.RDB.Set(ctx, rdbKey, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}

func (h *Here) ScopeLogical() (int64, error) {
	slog := log.With().Str("id", h.ID()).Logger()
	rdbKey := "frontend_discord_scope_here_logical_" + h.ID()

	scope, err := core.RDB.Get(ctx, rdbKey).Int64()
	if err != nil && err != redis.Nil {
		return -1, err
	}
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = getPlaceLogicalScope(h.ID(), h.ChannelID, h.GuildID)
	if err != nil {
		return -1, err
	}
	err = core.RDB.Set(ctx, rdbKey, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}
