package core

import (
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var RDB *redis.Client

// CacheScope returns the scope by looking it up in the cache, if it doesn't
// exist then it fetches it from the DB using getScope and then caches it. The
// key should be globally unique.
func CacheScope(key string, getScope func() (int64, error)) (int64, error) {
	slog := log.With().Str("key", key).Logger()

	scope, err := RDB.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return -1, err
	}
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = getScope()
	if err != nil {
		return -1, err
	}
	err = RDB.Set(ctx, key, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}
