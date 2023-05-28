package core

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	EventMessage      = make(chan *Message)
	EventMessageHooks = hooks[*Message]{}

	EventRedeemClaim      = make(chan *RedeemClaim)
	EventRedeemClaimHooks = hooks[*RedeemClaim]{}

	EventStreamOnline      = make(chan *StreamOnline)
	EventStreamOnlineHooks = hooks[*StreamOnline]{}

	EventStreamOffline      = make(chan *StreamOffline)
	EventStreamOfflineHooks = hooks[*StreamOffline]{}
)

type RedeemClaim struct {
	ID     string
	Input  string
	When   time.Time
	Author Author
	Here   Here
}

type StreamOnline struct {
	When time.Time
	Here Here
}

type StreamOffline struct {
	When time.Time
	Here Here
}

// EventLoop starts an infinite loop which handles all the events
func EventLoop() {
	for {
		select {
		case m := <-EventMessage:
			log.Debug().Msg("received message event")
			EventMessageHooks.Run(m)
			if _, err := m.CommandRun(); err != nil {
				log.Debug().Err(err).Send()
			}

		case redeem := <-EventRedeemClaim:
			log.Debug().Msg("received redeem claim event")
			EventRedeemClaimHooks.Run(redeem)

		case on := <-EventStreamOnline:
			log.Debug().Msg("received stream online event")
			here, err := on.Here.ScopeLogical()
			if err != nil {
				log.Error().Err(err).Msg("failed to get logical here")
				break
			}
			if err := streamOnline(here, on.When); err != nil {
				log.Error().Err(err).Msg("failed to save stream online status")
				break
			}
			EventStreamOnlineHooks.Run(on)

		case off := <-EventStreamOffline:
			log.Debug().Msg("received stream offline event")
			here, err := off.Here.ScopeLogical()
			if err != nil {
				log.Error().Err(err).Msg("failed to get logical here")
				break
			}
			if err := DB.PlaceSet("stream_offline_actual", here, off.When.UTC().Unix()); err != nil {
				log.Error().Err(err).Msg("failed save stream offline actual")
				break
			}
			EventStreamOfflineHooks.Run(off)
		}
	}
}

// streamOnline will save the timestamp of when the stream went online. It tries
// to filter shaky connections by giving a grace period of the stream going
// offline and online again (event multiple times), in which case the streams
// are viewed as one.
func streamOnline(place int64, when time.Time) error {
	// There are 2 kinds of values, actual and normalized. Actual keeps track of
	// online/offline events as they come in, without any filtering, normalized
	// makes sure that more than the specified grace period has passed between
	// the stream going down and up again.

	DB.Lock.Lock()
	defer DB.Lock.Unlock()

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	err = tx.PlaceSet("stream_online_actual", place, when.UTC().Unix())
	if err != nil {
		return err
	}

	offline, err := tx.PlaceGet("stream_offline_actual", place)
	if err != nil {
		return err
	}

	grace, err := tx.PlaceGet("stream_grace", place)
	if err != nil {
		return err
	}

	diff := when.Sub(time.Unix(offline.(int64), 0))
	if diff <= time.Duration(grace.(int64))*time.Second {
		log.Debug().
			Str("diff", diff.String()).
			Msg("stream online again within grace period")
		// in order to save the stream_online_actual value
		return tx.Commit()
	}

	offlinePrev, err := tx.PlaceGet("stream_offline_norm", place)
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm_prev", place, offlinePrev.(int64))
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm", place, offline.(int64))
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_online_norm", place, when.UTC().Unix())
	if err != nil {
		return err
	}
	return tx.Commit()
}

type hook[T any] struct {
	ID  int
	Run func(T)
}

// Hooks are a list of functions that are applied one-by-one to incoming
// events. All operations are thread safe.
type hooks[T any] struct {
	lock  sync.RWMutex
	hooks []hook[T]

	// Keeps track of the number of hooks added, is incremented every time a
	// new hook is added, does not get decreased if a hook is removed. Used as
	// a hook ID.
	total int
}

// Register returns the hook's id which can be used to delete the hook by
// calling the Delete method.
func (hs *hooks[T]) Register(f func(T)) int {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	hs.total++
	h := hook[T]{
		ID:  hs.total,
		Run: f,
	}
	hs.hooks = append(hs.hooks, h)

	return hs.total
}

// Delete will delete the hook with the given id. If the hook doesn't exist then
// nothing happens.
func (hs *hooks[T]) Delete(id int) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	for i, h := range hs.hooks {
		if h.ID == id {
			hs.hooks = append(hs.hooks[:i], hs.hooks[i+1:]...)
			return
		}
	}
}

func (hs *hooks[T]) Run(arg T) {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	for _, h := range hs.hooks {
		h.Run(arg)
	}
}
