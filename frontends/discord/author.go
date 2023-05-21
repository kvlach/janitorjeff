package discord

import (
	"errors"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

func getAuthorScope(authorID string) (int64, error) {
	rdbKey := "frontend_discord_scope_author_" + authorID

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbGetPersonScope(authorID)
	})
}

// AuthorMessage implements the core.Author interface
type AuthorMessage struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member
}

func (a *AuthorMessage) ID() (string, error) {
	if a.Author != nil && a.Author.ID != "" {
		return a.Author.ID, nil
	}
	if a.Member != nil && a.Member.User != nil && a.Member.User.ID != "" {
		return a.Member.User.ID, nil
	}
	return "", errors.New("can't figure out author id")
}

func (a *AuthorMessage) Name() (string, error) {
	if a.Author != nil && a.Author.Username != "" {
		return a.Author.Username, nil
	}
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}
	id, err := a.ID()
	if err != nil {
		return "", err
	}
	user, err := Session.User(id)
	if err != nil {
		return "", err
	}
	a.Author = user
	return user.Username, nil
}

func (a *AuthorMessage) DisplayName() (string, error) {
	var mem *dg.Member

	aid, err := a.ID()
	if err != nil {
		return "", err
	}

	if a.Member != nil {
		mem = a.Member
	} else {
		if a.GuildID == "" {
			return "", errors.New("need guild id")
		}
		mem, err = Session.GuildMember(a.GuildID, aid)
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
	return isAdmin(a.GuildID, aid), nil
}

func (a *AuthorMessage) Mod() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.GuildID, aid), nil
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
