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
func CreateContext(person, place int64, msgID string) (*core.Message, error) {
	frontend, err := core.DB.ScopeFrontend(place)
	if err != nil {
		return nil, err
	}

	var client core.Messenger

	switch frontend {
	case int64(discord.Frontend.Type()):
		client, err = discord.CreateClient(person, place, msgID)
	case int64(twitch.Frontend.Type()):
		client, err = twitch.CreateClient(person, place)
	default:
		return nil, fmt.Errorf("frontend with id '%d' is not supported", frontend)
	}

	if err != nil {
		return nil, err
	}

	return client.Parse()
}
