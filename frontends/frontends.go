package frontends

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"
)

var Frontends = core.Frontenders{
	discord.Frontend,
	twitch.Frontend,
}
