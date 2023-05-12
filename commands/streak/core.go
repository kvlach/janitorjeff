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

var (
	ErrRedeemNotSet = errors.New("the streak redeem has not been set")
	ErrIgnore       = errors.New("stream online within grace period, do nothing")
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

func Appearance(person, place int64, when time.Time) (int64, error) {
	core.DB.Lock.Lock()
	defer core.DB.Lock.Unlock()

	tx, err := core.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	prev, err := tx.PersonGet("cmd_streak_last", person, place)
	if err != nil {
		return 0, err
	}
	online, err := tx.PlaceGet("cmd_streak_online_norm", place)
	if err != nil {
		return -1, err
	}
	offline, err := tx.PlaceGet("cmd_streak_offline_norm", place)
	if err != nil {
		return 0, err
	}

	// This will get triggered in the following scenario:
	//
	//     online -> offline -> online -> redeem
	//
	// In which case the streak counter gets reset to 0 as the person didn't
	// show up in the previous stream
	if offline.(int64) > prev.(int64) {
		err = tx.PersonSet("cmd_streak_num", person, place, 0)
		if err != nil {
			return 0, err
		}
	}

	// This will get triggered in the following scenario:
	//
	//     online -> redeem -> offline -> online (within grace) -> redeem
	//
	// In which case the streak doesn't get incremented as this is considered
	// one stream.
	if prev.(int64) >= online.(int64) {
		return 0, ErrIgnore
	}

	err = tx.PersonSet("cmd_streak_last", person, place, when.UTC().Unix())
	if err != nil {
		return 0, err
	}
	streak, err := tx.PersonGet("cmd_streak_num", person, place)
	if err != nil {
		return -1, err
	}
	err = tx.PersonSet("cmd_streak_num", person, place, streak.(int64)+1)
	if err != nil {
		return -1, err
	}
	return streak.(int64) + 1, tx.Commit()
}

// Online will save the timestamp of when the stream went online. It tries to
// filter shaky connections by giving a grace period of the stream going offline
// and online again (event multiple times), in which case the streams are viewed
// as one.
func Online(place int64, when time.Time) error {
	// There are 2 kinds of values, actual and normalized. Actual keeps track of
	// online/offline events as they come in, without any filtering, normalized
	// makes sure that more than the specified grace period has passed between
	// the stream going down and up again.

	core.DB.Lock.Lock()
	defer core.DB.Lock.Unlock()

	tx, err := core.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.PlaceSet("cmd_streak_online_actual", place, when.UTC().Unix())
	if err != nil {
		return err
	}

	offline, err := tx.PlaceGet("cmd_streak_offline_actual", place)
	if err != nil {
		return err
	}

	grace, err := tx.PlaceGet("cmd_streak_grace", place)
	if err != nil {
		return err
	}

	diff := when.Sub(time.Unix(offline.(int64), 0))
	if diff <= time.Duration(grace.(int64))*time.Second {
		log.Debug().
			Str("diff", diff.String()).
			Msg("stream online again within grace period")
		return ErrIgnore
	}

	err = tx.PlaceSet("cmd_streak_offline_norm", place, offline.(int64))
	if err != nil {
		return err
	}
	return tx.PlaceSet("cmd_streak_online_norm", place, when.UTC().Unix())
}

func Offline(place int64, when time.Time) error {
	return core.DB.PlaceSet("cmd_streak_offline_actual", place, when.UTC().Unix())
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

func Get(person, place int64) (int64, error) {
	streak, err := core.DB.PersonGet("cmd_streak_num", person, place)
	if err != nil {
		return 0, err
	}
	return streak.(int64), nil
}

func GraceGet(place int64) (time.Duration, error) {
	grace, err := core.DB.PlaceGet("cmd_streak_grace", place)
	if err != nil {
		return 0, err
	}
	return time.Duration(grace.(int64)) * time.Second, nil
}

func GraceSet(place int64, grace time.Duration) error {
	return core.DB.PlaceSet("cmd_streak_grace", place, int(grace.Seconds()))
}
