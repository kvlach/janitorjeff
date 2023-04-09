package core

import (
	"sync"
)

// Hooks are a list of functions that are applied one-by-one to incoming
// messages. All operations are thread safe.
var Hooks = hooks{}

type hook struct {
	ID  int
	Run func(m *Message)
}

type hooks struct {
	lock  sync.RWMutex
	hooks []hook

	// Keeps track of the number of hooks added, is incremented every time a
	// new hook is added, does not get decreased if a hook is removed. Used as
	// a hook ID.
	total int
}

// Register returns the hook's id which can be used to delete the hook by
// calling the Delete method.
func (hs *hooks) Register(f func(*Message)) int {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	hs.total++
	h := hook{
		ID:  hs.total,
		Run: f,
	}
	hs.hooks = append(hs.hooks, h)

	return hs.total
}

// Delete will delete the hook with the given id. If the hook doesn't exist then
// nothing happens.
func (hs *hooks) Delete(id int) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	for i, h := range hs.hooks {
		if h.ID == id {
			hs.hooks = append(hs.hooks[:i], hs.hooks[i+1:]...)
			return
		}
	}
}

// Get returns the list of hooks.
func (hs *hooks) Get() []hook {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	return hs.hooks
}
