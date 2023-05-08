package discord

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

func getAuthorScope(authorID string) (int64, error) {
	rdbKey := "frontend_discord_scope_author_" + authorID

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbGetPersonScope(authorID)
	})
}

// Implement the core.Author interface for normal messages
type AuthorMessage struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member
}

func (a *AuthorMessage) ID() (string, error) {
	return a.Author.ID, nil
}

func (a *AuthorMessage) Name() (string, error) {
	return a.Author.Username, nil
}

func (a *AuthorMessage) DisplayName() (string, error) {
	return getDisplayName(a.Member, a.Author), nil
}

func (a *AuthorMessage) Mention() (string, error) {
	return a.Author.Mention(), nil
}

func (a *AuthorMessage) BotAdmin() (bool, error) {
	return isBotAdmin(a.Author.ID), nil
}

func (a *AuthorMessage) Admin() (bool, error) {
	return isAdmin(a.GuildID, a.Author.ID), nil
}

func (a *AuthorMessage) Mod() (bool, error) {
	return isMod(a.GuildID, a.Author.ID), nil
}

func (a *AuthorMessage) Subscriber() (bool, error) {
	return false, nil
}

func (a *AuthorMessage) Scope() (int64, error) {
	aid, err := a.ID()
	if err != nil {
		return 0, err
	}
	return getAuthorScope(aid)
}

// Implement the core.Author interface for interactions
type AuthorInteraction struct {
	GuildID string
	Member  *dg.Member
	User    *dg.User
}

func (a *AuthorInteraction) ID() (string, error) {
	if a.Member != nil {
		return a.Member.User.ID, nil
	}
	return a.User.ID, nil
}

func (a *AuthorInteraction) Name() (string, error) {
	if a.Member != nil {
		return a.Member.User.Username, nil
	}
	return a.User.Username, nil
}

func (a *AuthorInteraction) DisplayName() (string, error) {
	if a.Member != nil {
		return a.Member.User.Username, nil
	}
	return a.User.Username, nil
}

func (a *AuthorInteraction) Mention() (string, error) {
	if a.Member != nil {
		return a.Member.Mention(), nil
	}
	return a.User.Mention(), nil
}

func (a *AuthorInteraction) BotAdmin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isBotAdmin(aid), nil
}

func (a *AuthorInteraction) Admin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isAdmin(a.GuildID, aid), nil
}

func (a *AuthorInteraction) Mod() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.GuildID, aid), nil
}

func (a *AuthorInteraction) Subscriber() (bool, error) {
	return false, nil
}

func (a *AuthorInteraction) Scope() (int64, error) {
	aid, err := a.ID()
	if err != nil {
		return 0, err
	}
	return getAuthorScope(aid)
}
