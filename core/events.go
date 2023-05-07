package core

import (
	"sync"

	"github.com/rs/zerolog/log"
)

var (
	EventMessage      = make(chan *Message)
	EventMessageHooks = hooks[*Message]{}

	EventRedeemClaim      = make(chan string)
	EventRedeemClaimHooks = hooks[string]{}

	EventStreamOnline      = make(chan struct{})
	EventStreamOnlineHooks = hooks[struct{}]{}

	EventStreamOffline      = make(chan struct{})
	EventStreamOfflineHooks = hooks[struct{}]{}
)

func init() {
	go Events()
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

func Events() {
	for {
		select {
		case m := <-EventMessage:
			EventMessageHooks.Run(m)
			if _, err := m.CommandRun(); err != nil {
				log.Debug().Err(err).Send()
			}

		case redeemUUID := <-EventRedeemClaim:
			EventRedeemClaimHooks.Run(redeemUUID)

		case <-EventStreamOnline:
			EventStreamOnlineHooks.Run(struct{}{})

		case <-EventStreamOffline:
			EventStreamOfflineHooks.Run(struct{}{})
		}
	}
}
