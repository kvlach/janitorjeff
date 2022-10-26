package core

import (
	"sync"

	dg "github.com/bwmarrin/discordgo"
)

type DiscordVars struct {
	Client        *dg.Session
	EmbedColor    int
	EmbedErrColor int
	Admins        []string
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

type AllCommands struct {
	Normal   Commands
	Advanced Commands
	Admin    Commands
}

type Prefixes struct {
	Admin  []Prefix
	Others []Prefix
}

type GlobalVars struct {
	Commands AllCommands
	DB       *DB
	Host     string
	Hooks    Hooks
	Prefixes Prefixes

	// Platform Specific
	Discord *DiscordVars
	Twitch  *TwitchVars
}

var Globals *GlobalVars

func GlobalsInit(g *GlobalVars) {
	Globals = g
}
