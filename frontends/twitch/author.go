package twitch

import (
	tirc "github.com/gempir/go-twitch-irc/v2"
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

func (a Author) Scope() (int64, error) {
	return dbAddChannel(a.ID(), a.User, nil)
}

func (a Author) ScopeExact() (int64, error) {
	return a.Scope()
}

func (a Author) ScopeLogical() (int64, error) {
	return a.Scope()
}
