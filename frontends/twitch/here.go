package twitch

import (
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
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
	slog := log.With().Str("here", h.ID()).Logger()
	rdbKey := "frontend_twitch_scope_" + h.ID()

	scope, err := core.RDB.Get(ctx, rdbKey).Int64()
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = dbAddChannelSimple(h.ID(), h.Name())
	if err != nil {
		return -1, err
	}
	err = core.RDB.Set(ctx, rdbKey, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}

func (h Here) ScopeExact() (int64, error) {
	return h.Scope()
}

func (h Here) ScopeLogical() (int64, error) {
	return h.Scope()
}
