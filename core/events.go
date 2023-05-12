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

func init() {
	go Events()
}

func Events() {
	for {
		select {
		case m := <-EventMessage:
			EventMessageHooks.Run(m)
			if _, err := m.CommandRun(); err != nil {
				log.Debug().Err(err).Send()
			}

		case redeem := <-EventRedeemClaim:
			EventRedeemClaimHooks.Run(redeem)

		case on := <-EventStreamOnline:
			EventStreamOnlineHooks.Run(on)

		case off := <-EventStreamOffline:
			EventStreamOfflineHooks.Run(off)
		}
	}
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
