package core

import (
	"git.slowtyper.com/slowtyper/janitorjeff/sqldb"
)

type DiscordVars struct {
	EmbedColor int
}

type TwitchVars struct {
	ClientID     string
	ClientSecret string
}

type GlobalVars struct {
	Commands  Commands
	DB        *sqldb.DB
	Host      string
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
