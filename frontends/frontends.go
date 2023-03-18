package frontends

import (
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"
	"github.com/janitorjeff/jeff-bot/frontends/twitch"
)

var Frontends = core.Frontenders{
	discord.Frontend,
	twitch.Frontend,
}
