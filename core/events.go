package core

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

type Event[T any] interface {
	*Message | *RedeemClaim | *StreamOnline | *StreamOffline

	Hooks() *Hooks[T]
}

var (
	EventMessage        = make(chan *Message)
	EventMessageHooks   = HooksNew[*Message](20)
	eventMessageCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jeff_event_message_total",
		Help: "Total number of received message events.",
	}, []string{"frontend", "person", "place"})

	EventRedeemClaim        = make(chan *RedeemClaim)
	EventRedeemClaimHooks   = HooksNew[*RedeemClaim](5)
	eventRedeemClaimCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jeff_event_redeem_claim_total",
		Help: "Total number of received redeem claim events.",
	})

	EventStreamOnline        = make(chan *StreamOnline)
	EventStreamOnlineHooks   = HooksNew[*StreamOnline](5)
	eventStreamOnlineCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jeff_event_stream_online_total",
		Help: "Total number of received stream online events.",
	})

	EventStreamOffline        = make(chan *StreamOffline)
	EventStreamOfflineHooks   = HooksNew[*StreamOffline](5)
	eventStreamOfflineCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jeff_event_stream_offline_total",
		Help: "Total number of received stream offline events.",
	})
)

func (m *Message) Hooks() *Hooks[*Message] {
	return EventMessageHooks
}

type RedeemClaim struct {
	ID     string
	Input  string
	When   time.Time
	Author Personifier
	Here   Placer
}

func (rc *RedeemClaim) Hooks() *Hooks[*RedeemClaim] {
	return EventRedeemClaimHooks
}

type StreamOnline struct {
	When time.Time
	Here Placer
}

func (son *StreamOnline) Hooks() *Hooks[*StreamOnline] {
	return EventStreamOnlineHooks
}

type StreamOffline struct {
	When time.Time
	Here Placer
}

func (son *StreamOffline) Hooks() *Hooks[*StreamOffline] {
	return EventStreamOfflineHooks
}

// EventLoop starts an infinite loop which handles all incoming events. It's
// possible to have multiple instances running in separate goroutines (in order
// for the bot not to lag when handling an event that takes longer than
// virtually instantly); Golang guarantees that only one of the receivers will
// receive the channel data.
func EventLoop() {
	for {
		select {
		case m := <-EventMessage:
			person, err := m.Author.Scope()
			if err != nil {
				log.Error().Err(err)
			}
			place, err := m.Here.ScopeLogical()
			if err != nil {
				log.Error().Err(err)
			}
			eventMessageCounter.
				With(prometheus.Labels{
					"frontend": m.Frontend.Name(),
					"person":   strconv.FormatInt(person, 10),
					"place":    strconv.FormatInt(place, 10),
				}).
				Inc()
			log.Debug().
				Str("id", m.ID).
				Str("raw", m.Raw).
				Interface("frontend", m.Frontend.Type()).
				Msg("received message event")
			EventMessageHooks.Run(m)
			if _, err := m.CommandRun(); err != nil {
				log.Debug().
					Err(err).
					Interface("command", m.Command).
					Msg("got error while running command")
			}

		case rc := <-EventRedeemClaim:
			eventRedeemClaimCounter.Inc()
			log.Debug().
				Str("id", rc.ID).
				Str("input", rc.Input).
				Str("when", rc.When.String()).
				Msg("received redeem claim event")
			EventRedeemClaimHooks.Run(rc)

		case on := <-EventStreamOnline:
			eventStreamOnlineCounter.Inc()
			log.Debug().
				Str("when", on.When.String()).
				Msg("received stream online event")
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
			eventStreamOfflineCounter.Inc()
			log.Debug().
				Str("when", off.When.String()).
				Msg("received stream offline event")
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

	offline, err := tx.PlaceGet("stream_offline_actual", place).Time()
	if err != nil {
		return err
	}

	grace, err := tx.PlaceGet("stream_grace", place).Duration()
	if err != nil {
		return err
	}

	diff := when.Sub(offline)
	if diff <= grace {
		log.Debug().
			Str("diff", diff.String()).
			Msg("stream online again within grace period")
		// in order to save the stream_online_actual value
		return tx.Commit()
	}

	offlinePrev, err := tx.PlaceGet("stream_offline_norm", place).Int64()
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm_prev", place, offlinePrev)
	if err != nil {
		return err
	}
	err = tx.PlaceSet("stream_offline_norm", place, offline.Unix())
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

type hookData[T any] struct {
	Hook hook[T]
	Arg  T
}

// Hooks are a list of functions that are applied one-by-one to incoming
// events. All operations are thread safe.
type Hooks[T any] struct {
	lock  sync.RWMutex
	hooks []hook[T]
	ch    chan hookData[T]

	// Tracks the number of added hooks, is incremented every time a new hook is
	// added, does not get decreased if a hook is removed.
	// Used as a hook ID.
	total int
}

// HooksNew generates a new Hooks object. Spawns n number of receiver functions
// in their own goroutines.
func HooksNew[T any](n int) *Hooks[T] {
	hs := &Hooks[T]{}
	hs.ch = make(chan hookData[T])

	for i := 0; i < n; i++ {
		go func() {
			for {
				data := <-hs.ch
				log.Debug().
					Int("hook-id", data.Hook.ID).
					Interface("arg", data.Arg).
					Msg("received data for hook")
				data.Hook.Run(data.Arg)
			}
		}()
	}

	return hs
}

// Register returns the hook's id which can be used to delete the hook by
// calling the Delete method.
func (hs *Hooks[T]) Register(f func(T)) int {
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

// Delete will delete the hook with the given id. If the hook doesn't exist,
// then nothing happens.
func (hs *Hooks[T]) Delete(id int) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	for i, h := range hs.hooks {
		if h.ID == id {
			hs.hooks = append(hs.hooks[:i], hs.hooks[i+1:]...)
			return
		}
	}
}

func (hs *Hooks[T]) Run(arg T) {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	for _, h := range hs.hooks {
		hs.ch <- hookData[T]{h, arg}
	}
}

// EventAwait monitors incoming events until check is true or until timeout. If
// nothing is matched, then the returned object will be the default value of the
// type.
func EventAwait[T Event[T]](timeout time.Duration, check func(T) bool) T {
	found := make(chan struct{})

	var t T
	h := t.Hooks()
	id := h.Register(func(candidate T) {
		if check(candidate) {
			t = candidate
			found <- struct{}{}
		}
	})

	select {
	case <-found:
		break
	case <-time.After(timeout):
		break
	}

	h.Delete(id)
	return t
}
