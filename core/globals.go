package core

import (
	"sync"

	"git.slowtyper.com/slowtyper/janitorjeff/sqldb"
)

type DiscordVars struct {
	EmbedColor    int
	EmbedErrColor int
}

type TwitchVars struct {
	ClientID     string
	ClientSecret string
}

type Hooks struct {
	lock  sync.RWMutex
	hooks []func(*Message)
}

func (h *Hooks) Register(f func(*Message)) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.hooks = append(h.hooks, f)
}

func (h *Hooks) Get() []func(*Message) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.hooks
}

type GlobalVars struct {
	Commands  Commands
	DB        *sqldb.DB
	Host      string
	Hooks     Hooks
	Prefixes_ []string

	// Platform Specific
	Discord *DiscordVars
	Twitch  *TwitchVars
}

var Globals *GlobalVars

func GlobalsInit(g *GlobalVars) {
	Globals = g
}

func (g *GlobalVars) Prefixes() []string {
	// Creates a copy instead of a reference to make modifying the prefixes,
	// e.g. for renering easier, since otherwise it would modify the global
	// variable.
	prefixes := make([]string, len(g.Prefixes_))
	copy(prefixes, g.Prefixes_)
	return prefixes
}
