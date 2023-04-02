package twitch

import (
	"github.com/janitorjeff/jeff-bot/core"

	tirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Author implements both the core.Author and core.Here interfaces since users
// and channels are the same thing on twitch.
type Author struct {
	User tirc.User
}

func (a Author) ID() string {
	return a.User.ID
}

func (a Author) Name() string {
	return a.User.Name
}

func (a Author) DisplayName() string {
	return a.User.DisplayName
}

func (a Author) Mention() string {
	return "@" + a.DisplayName()
}

func (a Author) BotAdmin() bool {
	return false
}

func (a Author) Admin() bool {
	_, ok := a.User.Badges["broadcaster"]
	return ok
}

func (a Author) Mod() bool {
	if a.Admin() {
		return true
	}
	_, ok := a.User.Badges["moderator"]
	return ok
}

func (a Author) Subscriber() bool {
	if _, ok := a.User.Badges["subscriber"]; ok {
		return true
	}
	_, ok := a.User.Badges["founder"]
	return ok
}

func (a Author) Scope() (int64, error) {
	slog := log.With().Str("author", a.ID()).Logger()
	rdbKey := "frontend_twitch_scope_" + a.ID()

	scope, err := core.RDB.Get(ctx, rdbKey).Int64()
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = dbAddChannel(a.ID(), a.User, nil)
	if err != nil {
		return -1, err
	}
	err = core.RDB.Set(ctx, rdbKey, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}

func (a Author) ScopeExact() (int64, error) {
	return a.Scope()
}

func (a Author) ScopeLogical() (int64, error) {
	return a.Scope()
}
