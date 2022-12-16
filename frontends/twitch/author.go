package twitch

import (
	tirc "github.com/gempir/go-twitch-irc/v2"
)

// Implement the core.Author interface

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

func (a Author) Scope() (int64, error) {
	return dbAddChannel(a.ID(), a.User, nil)
}
