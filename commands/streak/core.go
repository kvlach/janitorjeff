package streak

import (
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/nicklaw5/helix"
)

func On(h *twitch.Helix, broadcasterID string) error {
	if err := h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOnline); err != nil {
		return err
	}
	return h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOffline)
}

func Off(h *twitch.Helix, broadcasterID string) error {
	return nil
}
