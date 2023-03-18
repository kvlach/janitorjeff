package frontends

import (
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

	// This always matches something
	for _, frontend := range Frontends {
		if frontend.Type() == core.FrontendType(frontendType) {
			f = frontend
			break
		}
	}

	return f.CreateMessage(person, place, msgID)
}
