package core

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Event[T any] interface {
	*Message | *RedeemClaim | *StreamOnline | *StreamOffline

	Hooks() *Hooks[T]

	Handler()
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
			m.Handler()
		case rc := <-EventRedeemClaim:
			rc.Handler()
		case son := <-EventStreamOnline:
			son.Handler()
		case soff := <-EventStreamOffline:
			soff.Handler()
		}
	}
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
