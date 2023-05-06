package streak

import (
	"database/sql"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

func On(h *twitch.Helix, place int64, broadcasterID string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	onlineSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOnline)
	if err != nil {
		return err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, onlineSubID); err != nil {
		return err
	}

	offlineSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOffline)
	if err != nil {
		return err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, offlineSubID); err != nil {
		return err
	}

	redeemsSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd)
	if err != nil {
		return err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, redeemsSubID); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO cmd_streak_twitch_events(place, event_online, event_offline, event_redeem)
		VALUES ($1, $2, $3, $4)`, place, onlineSubID, offlineSubID, redeemsSubID)

	log.Debug().
		Err(err).
		Str("online", onlineSubID).
		Str("offline", offlineSubID).
		Str("redeem", redeemsSubID).
		Msg("saved subscription ids")

	if err != nil {
		return err
	}

	return tx.Commit()
}

func Off(h *twitch.Helix, place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	var onlineSubID, offlineSubID, redeemsSubID string
	err = tx.QueryRow(`
		SELECT event_online, event_offline, event_redeem
		FROM cmd_streak_twitch_events
		WHERE place = $1`, place).Scan(&onlineSubID, &offlineSubID, &redeemsSubID)

	log.Debug().
		Err(err).
		Str("online", onlineSubID).
		Str("offline", offlineSubID).
		Str("redeem", redeemsSubID).
		Msg("got subscription ids")

	if err != nil {
		return err
	}

	if err := h.DeleteSubscription(onlineSubID); err != nil {
		return err
	}
	if err := twitch.DeleteEventSubSubscriptionID(tx, onlineSubID); err != nil {
		return err
	}

	if err := h.DeleteSubscription(offlineSubID); err != nil {
		return err
	}
	if err := twitch.DeleteEventSubSubscriptionID(tx, offlineSubID); err != nil {
		return err
	}

	if err := h.DeleteSubscription(redeemsSubID); err != nil {
		return err
	}
	if err := twitch.DeleteEventSubSubscriptionID(tx, redeemsSubID); err != nil {
		return err
	}

	return tx.Commit()
}
