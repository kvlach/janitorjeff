package frontends

import (
	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"
	"github.com/kvlach/janitorjeff/frontends/twitch"
)

var Frontends = core.Frontenders{
	discord.Frontend,
	twitch.Frontend,
}
