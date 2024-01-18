package discord

import (
	"errors"
	"sync"

	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func getAuthorScope(authorID string) (int64, error) {
	rdbKey := "frontend_discord_scope_author_" + authorID

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbGetPersonScope(authorID)
	})
}

// AuthorMessage implements the core.Personifier interface
type AuthorMessage struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member

	mu    sync.Mutex
	scope int64
}

func (a *AuthorMessage) id() (string, error) {
	if a.Author != nil && a.Author.ID != "" {
		return a.Author.ID, nil
	}
	if a.Member != nil && a.Member.User != nil && a.Member.User.ID != "" {
		return a.Member.User.ID, nil
	}
	return "", errors.New("can't figure out author id")
}

func (a *AuthorMessage) ID() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.id()
}

func (a *AuthorMessage) name() (string, error) {
	if a.Author != nil && a.Author.Username != "" {
		return a.Author.Username, nil
	}
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}

	id, err := a.id()
	if err != nil {
		return "", err
	}
	user, err := Client.Session.User(id)
	if err != nil {
		return "", err
	}
	a.Author = user
	return user.Username, nil
}

func (a *AuthorMessage) Name() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.name()
}

func (a *AuthorMessage) DisplayName() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var mem *dg.Member

	aid, err := a.id()
	if err != nil {
		return "", err
	}

	if a.Member != nil {
		mem = a.Member
	} else {
		if a.GuildID == "" {
			return "", errors.New("need guild id")
		}
		mem, err = Client.Member(a.GuildID, aid)
		a.Member = mem
		if err != nil {
			return "", err
		}
	}

	// Return the display name if it exists, otherwise return the username
	if mem.Nick != "" {
		return a.Member.Nick, nil
	} else {
		return a.Name()
	}
}

func (a *AuthorMessage) Mention() (string, error) {
	id, err := a.ID()
	if err != nil {
		return "", err
	}
	return "<@" + id + ">", nil
}

func (a *AuthorMessage) BotAdmin() (bool, error) {
	id, err := a.ID()
	if err != nil {
		return false, err
	}
	return isBotAdmin(id), nil
}

func (a *AuthorMessage) Admin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isAdmin(a.guildID, aid)
}

func (a *AuthorMessage) Moderator() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.guildID, aid)
}

func (a *AuthorMessage) Subscriber() (bool, error) {
	return false, nil
}

func (a *AuthorMessage) Scope() (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.scope != 0 {
		log.Debug().
			Int64("scope", a.scope).
			Msg("LOCAL: found cached scope")
		return a.scope, nil
	}

	aid, err := a.id()
	if err != nil {
		return 0, err
	}
	scope, err := getAuthorScope(aid)
	if err != nil {
		return 0, err
	}
	a.scope = scope
	return scope, nil
}

// AuthorInteraction implements the core.Personifier interface for interactions
type AuthorInteraction struct {
	GuildID string
	Member  *dg.Member
	User    *dg.User

	scope int64
}

func (a *AuthorInteraction) ID() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.ID != "" {
		return a.Member.User.ID, nil
	}
	if a.User != nil && a.User.ID != "" {
		return a.User.ID, nil
	}
	return "", errors.New("can't figure out author id")
}

func (a *AuthorInteraction) Name() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}
	if a.User != nil && a.User.Username != "" {
		return a.User.Username, nil
	}
	return "", errors.New("can't figure out author name")
}

func (a *AuthorInteraction) DisplayName() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}
	if a.User != nil && a.User.Username != "" {
		return a.User.Username, nil
	}
	return "", errors.New("can't figure out author display name")
}

func (a *AuthorInteraction) Mention() (string, error) {
	// Member.Mention() requires Member.User
	if a.Member != nil && a.Member.User != nil {
		return a.Member.Mention(), nil
	}
	if a.User != nil {
		return a.User.Mention(), nil
	}
	return "", errors.New("can't figure out author mention")
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
	return isAdmin(a.GuildID, aid)
}

func (a *AuthorInteraction) Moderator() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.GuildID, aid)
}

func (a *AuthorInteraction) Subscriber() (bool, error) {
	return false, nil
}

func (a *AuthorInteraction) Scope() (int64, error) {
	if a.scope != 0 {
		return a.scope, nil
	}
	aid, err := a.ID()
	if err != nil {
		return 0, err
	}
	scope, err := getAuthorScope(aid)
	if err != nil {
		return 0, err
	}
	a.scope = scope
	return scope, nil
}
