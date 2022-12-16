package discord

import (
	dg "github.com/bwmarrin/discordgo"
)

// Implement the core.Author interface for normal messages
type AuthorMessage struct {
	// cache scope
	scope int64

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

func (a *AuthorMessage) Scope() (int64, error) {
	if a.scope != 0 {
		return a.scope, nil
	}
	author, err := getPersonScope(a.ID())
	if err != nil {
		return -1, err
	}
	a.scope = author
	return author, nil
}

// Implement the core.Author interface for interactions
type AuthorInteraction struct {
	//cache scope
	scope int64

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

func (a *AuthorInteraction) Scope() (int64, error) {
	if a.scope != 0 {
		return a.scope, nil
	}
	author, err := getPersonScope(a.ID())
	if err != nil {
		return -1, err
	}
	a.scope = author
	return author, nil
}
