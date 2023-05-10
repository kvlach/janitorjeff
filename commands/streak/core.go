package streak

import (
	"database/sql"
	"errors"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/google/uuid"
	"github.com/nicklaw5/helix/v2"
	"github.com/rs/zerolog/log"
)

var ErrRedeemNotSet = errors.New("the streak redeem has not been set")

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

func Reset(person, place int64) error {
	return core.DB.PersonSet("cmd_streak_num", person, place, 0)
}

func Appearance(person, place int64) error {
	core.DB.Lock.Lock()
	defer core.DB.Lock.Unlock()

	tx, err := core.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	online, err := tx.PlaceGet("cmd_streak_online", place)
	if err != nil {
		return err
	}
	offline, err := tx.PlaceGet("cmd_streak_offline", place)
	if err != nil {
		return err
	}
	diff := time.Unix(online.(int64), 0).Sub(time.Unix(offline.(int64), 0))
	// Don't increment the streak if the stream went down and up within a 30-min
	// period as in that case it was probably caused by technical difficulties,
	// and it's not a "different" stream.
	if diff < 30*time.Minute {
		log.Debug().
			Interface("diff", diff).
			Msg("stream online again within grace period")
		return nil
	}

	streak, err := tx.PersonGet("cmd_streak_num", person, place)
	if err != nil {
		return err
	}
	err = tx.PersonSet("cmd_streak_num", person, place, streak.(int64)+1)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func Online(place int64, when time.Time) error {
	return core.DB.PlaceSet("cmd_streak_online", place, when.UTC().Unix())
}

func Offline(place int64, when time.Time) error {
	return core.DB.PlaceSet("cmd_streak_offline", place, when.UTC().Unix())
}

func RedeemSet(place int64, id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return core.DB.PlaceSet("cmd_streak_redeem", place, u)
}

func RedeemGet(place int64) (uuid.UUID, error, error) {
	id, err := core.DB.PlaceGet("cmd_streak_redeem", place)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	if id == nil {
		return uuid.UUID{}, ErrRedeemNotSet, nil
	}
	u, err := uuid.Parse(string(id.([]uint8)))
	return u, nil, err
}
