package discord

import (
	"github.com/janitorjeff/jeff-bot/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func getAuthorScope(authorID string) (int64, error) {
	slog := log.With().Str("uid", authorID).Logger()
	rdbKey := "frontend_discord_scope_author_" + authorID

	scope, err := core.RDB.Get(ctx, rdbKey).Int64()
	if err != redis.Nil {
		slog.Debug().Int64("scope", scope).Msg("CACHE: found scope")
		return scope, nil
	}

	scope, err = dbGetPersonScope(authorID)
	if err != nil {
		return -1, err
	}
	err = core.RDB.Set(ctx, rdbKey, scope, 0).Err()
	slog.Debug().Err(err).Int64("scope", scope).Msg("CACHE: cached scope")
	return scope, err
}

// Implement the core.Author interface for normal messages
type AuthorMessage struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member
}

func (a *AuthorMessage) ID() string {
	return a.Author.ID
}

func (a *AuthorMessage) Name() string {
	return a.Author.Username
}

func (a *AuthorMessage) DisplayName() string {
	return getDisplayName(a.Member, a.Author)
}

func (a *AuthorMessage) Mention() string {
	return a.Author.Mention()
}

func (a *AuthorMessage) BotAdmin() bool {
	return isBotAdmin(a.Author.ID)
}

func (a *AuthorMessage) Admin() bool {
	return isAdmin(a.GuildID, a.Author.ID)
}

func (a *AuthorMessage) Mod() bool {
	return isMod(a.GuildID, a.Author.ID)
}

func (a *AuthorMessage) Subscriber() bool {
	return false
}

func (a *AuthorMessage) Scope() (int64, error) {
	return getAuthorScope(a.ID())
}

// Implement the core.Author interface for interactions
type AuthorInteraction struct {
	GuildID string
	Member  *dg.Member
	User    *dg.User
}

func (a *AuthorInteraction) ID() string {
	if a.Member != nil {
		return a.Member.User.ID
	}
	return a.User.ID
}

func (a *AuthorInteraction) Name() string {
	if a.Member != nil {
		return a.Member.User.Username
	}
	return a.User.Username
}

func (a *AuthorInteraction) DisplayName() string {
	if a.Member != nil {
		return a.Member.User.Username
	}
	return a.User.Username
}

func (a *AuthorInteraction) Mention() string {
	if a.Member != nil {
		return a.Member.Mention()
	}
	return a.User.Mention()
}

func (a *AuthorInteraction) BotAdmin() bool {
	return isBotAdmin(a.ID())
}

func (a *AuthorInteraction) Admin() bool {
	return isAdmin(a.GuildID, a.ID())
}

func (a *AuthorInteraction) Mod() bool {
	return isMod(a.GuildID, a.ID())
}

func (a *AuthorInteraction) Subscriber() bool {
	return false
}

func (a *AuthorInteraction) Scope() (int64, error) {
	return getAuthorScope(a.ID())
}
