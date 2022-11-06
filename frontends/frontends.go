package frontends

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"
)

const (
	Discord = discord.Type
	Twitch  = twitch.Type

	All = Discord | Twitch
)

// This is used to send messages that are not direct replies, e.g. reminders
func CreateContext(person, place int64) (*core.Message, error) {
	frontend, err := core.Globals.DB.ScopeFrontend(place)
	if err != nil {
		return nil, err
	}

	var client core.Messenger

	switch frontend {
	case Discord:
		client, err = discord.CreateClient(person, place)
	case Twitch:
		client, err = twitch.CreateClient(person, place)
	default:
		return nil, fmt.Errorf("frontend with id '%d' is not supported", frontend)
	}

	if err != nil {
		return nil, err
	}

	return client.Parse()
}
