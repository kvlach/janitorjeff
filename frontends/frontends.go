package frontends

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"
	"github.com/janitorjeff/jeff-bot/frontends/twitch"
)

var Frontends = []core.Frontender{
	discord.Frontend,
	twitch.Frontend,
}

// This is used to send messages that are not direct replies, e.g. reminders
func CreateMessage(person, place int64, msgID string) (*core.Message, error) {
	frontendType, err := core.DB.ScopeFrontend(place)
	if err != nil {
		return nil, err
	}

	var f core.Frontender

	switch frontendType {
	case int64(discord.Frontend.Type()):
		f = discord.Frontend
	case int64(twitch.Frontend.Type()):
		f = twitch.Frontend
	default:
		return nil, fmt.Errorf("frontend with id '%d' is not supported", frontendType)
	}

	return f.CreateMessage(person, place, msgID)
}
