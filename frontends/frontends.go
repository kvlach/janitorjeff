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

	for _, f := range Frontends {
		if f.Type() == core.FrontendType(frontendType) {
			return f.CreateMessage(person, place, msgID)
		}
	}

	return nil, fmt.Errorf("no frontend matched")
}
