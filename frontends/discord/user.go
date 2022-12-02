package discord

import (
	dg "github.com/bwmarrin/discordgo"
)

// Implement the core.User interface

type User struct {
	GuildID string
	Author  *dg.User
	Member  *dg.Member
}

func (u *User) ID() string {
	return u.Author.ID
}

func (u *User) Name() string {
	return u.Author.Username
}

func (u *User) DisplayName() string {
	return getDisplayName(u.Member, u.Author)
}

func (u *User) Mention() string {
	return u.Author.Mention()
}

func (u *User) BotAdmin() bool {
	return isBotAdmin(u.Author.ID)
}

func (u *User) Admin() bool {
	return isAdmin(Session, u.GuildID, u.Author.ID)
}

func (u *User) Mod() bool {
	return isMod(Session, u.GuildID, u.Author.ID)
}
