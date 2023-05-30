package streak

import (
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
	ErrAlreadyOn    = errors.New("streak tracking has already been turned on for this place")
)

func On(h *twitch.Helix, place int64, broadcasterID string) (error, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS (
		    SELECT 1 FROM cmd_streak_twitch_events
		    WHERE place = $1
		)`, place).Scan(&exists)
	if exists {
		return ErrAlreadyOn, nil
	}

	onlineSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOnline)
	if err != nil {
		return nil, err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, onlineSubID); err != nil {
		return nil, err
	}

	offlineSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeStreamOffline)
	if err != nil {
		return nil, err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, offlineSubID); err != nil {
		return nil, err
	}

	redeemsSubID, err := h.CreateSubscription(broadcasterID, helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd)
	if err != nil {
		return nil, err
	}
	if err := twitch.AddEventSubSubscriptionID(tx, redeemsSubID); err != nil {
		return nil, err
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
		return nil, err
	}

	return nil, tx.Commit()
}

func Off(h *twitch.Helix, place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

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
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	prev, err := tx.PersonGet("cmd_streak_last", person, place).Int64()
	if err != nil {
		return 0, err
	}
	offlinePrev, err := tx.PlaceGet("stream_offline_norm_prev", place).Int64()
	if err != nil {
		return 0, err
	}

	// This will get triggered in the following scenario:
	//
	//     online -> offline -> online -> redeem
	//
	// In which case the streak counter gets reset to 0 as the person didn't
	// show up in the previous stream
	if offlinePrev > prev {
		err = tx.PersonSet("cmd_streak_num", person, place, 0)
		if err != nil {
			return 0, err
		}
	}

	online, err := tx.PlaceGet("stream_online_norm", place).Int64()
	if err != nil {
		return -1, err
	}

	// This will get triggered in the following scenario:
	//
	//     online -> redeem -> offline -> online (within grace) -> redeem
	//
	// In which case the streak doesn't get incremented as this is considered
	// one stream.
	if prev >= online {
		return 0, ErrIgnore
	}

	err = tx.PersonSet("cmd_streak_last", person, place, when.UTC().Unix())
	if err != nil {
		return 0, err
	}
	streak, err := tx.PersonGet("cmd_streak_num", person, place).Int64()
	if err != nil {
		return -1, err
	}
	err = tx.PersonSet("cmd_streak_num", person, place, streak+1)
	if err != nil {
		return -1, err
	}
	return streak + 1, tx.Commit()
}

func RedeemSet(place int64, id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return core.DB.PlaceSet("cmd_streak_redeem", place, u)
}

func RedeemGet(place int64) (uuid.UUID, error, error) {
	id, isNil, err := core.DB.PlaceGet("cmd_streak_redeem", place).OptionalUUID()
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	if isNil {
		return uuid.UUID{}, ErrRedeemNotSet, nil
	}
	return id, nil, nil
}

func Get(person, place int64) (int64, error) {
	return core.DB.PersonGet("cmd_streak_num", person, place).Int64()
}

func Set(person, place int64, streak int) error {
	return core.DB.PersonSet("cmd_streak_num", person, place, streak)
}

func GraceGet(place int64) (time.Duration, error) {
	return core.DB.PlaceGet("stream_grace", place).Duration()
}

func GraceSet(place int64, grace time.Duration) error {
	return core.DB.PlaceSet("stream_grace", place, int(grace.Seconds()))
}
