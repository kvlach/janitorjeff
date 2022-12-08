package discord

import (
	dg "github.com/bwmarrin/discordgo"
)

// Implement the core.User interface for normal messages
type UserMessage struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member
}

func (u *UserMessage) ID() string {
	return u.Author.ID
}

func (u *UserMessage) Name() string {
	return u.Author.Username
}

func (u *UserMessage) DisplayName() string {
	return getDisplayName(u.Member, u.Author)
}

func (u *UserMessage) Mention() string {
	return u.Author.Mention()
}

func (u *UserMessage) BotAdmin() bool {
	return isBotAdmin(u.Author.ID)
}

func (u *UserMessage) Admin() bool {
	return isAdmin(u.GuildID, u.Author.ID)
}

func (u *UserMessage) Mod() bool {
	return isMod(u.GuildID, u.Author.ID)
}

// Implement the core.User interface for interactions
type UserInteraction struct {
	GuildID string
	Member  *dg.Member
	User    *dg.User
}

func (u *UserInteraction) ID() string {
	if u.Member != nil {
		return u.Member.User.ID
	}
	return u.User.ID
}

func (u *UserInteraction) Name() string {
	if u.Member != nil {
		return u.Member.User.Username
	}
	return u.User.Username
}

func (u *UserInteraction) DisplayName() string {
	if u.Member != nil {
		return u.Member.User.Username
	}
	return u.User.Username
}

func (u *UserInteraction) Mention() string {
	if u.Member != nil {
		return u.Member.Mention()
	}
	return u.User.Mention()
}

func (u *UserInteraction) BotAdmin() bool {
	return isBotAdmin(u.ID())
}

func (u *UserInteraction) Admin() bool {
	return isAdmin(u.GuildID, u.ID())
}

func (u *UserInteraction) Mod() bool {
	return isMod(u.GuildID, u.ID())
}
