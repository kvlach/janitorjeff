package twitch

import (
	tirc "github.com/gempir/go-twitch-irc/v2"
)

// Implement the core.User interface

type User struct {
	User tirc.User
}

func (u User) ID() string {
	return u.User.ID
}

func (u User) Name() string {
	return u.User.Name
}

func (u User) DisplayName() string {
	return u.User.DisplayName
}

func (u User) Mention() string {
	return "@" + u.DisplayName()
}

func (u User) BotAdmin() bool {
	return false
}

func (u User) Admin() bool {
	_, ok := u.User.Badges["broadcaster"]
	return ok
}

func (u User) Mod() bool {
	if u.Admin() {
		return true
	}
	_, ok := u.User.Badges["moderator"]
	return ok
}
