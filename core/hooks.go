package core

import (
	"sync"
)

var Hooks = hooks{}

type hook struct {
	id  int
	run func(m *Message)
}

type hooks struct {
	lock  sync.RWMutex
	hooks []hook

	// Keeps track of the number of hooks added, is incremented every time a
	// new hook is added, does not get decreased if a hook is removed. Used as
	// a hook ID.
	total int
}

// Returns the hook's id which is used to delete the hook.
func (hs *hooks) Register(f func(*Message)) int {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	hs.total++
	h := hook{
		id:  hs.total,
		run: f,
	}
	hs.hooks = append(hs.hooks, h)

	return hs.total
}

// Deletes the hook with the given id. If the hook doesn't exist then nothing
// happens.
func (hs *hooks) Delete(id int) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	for i, h := range hs.hooks {
		if h.id == id {
			hs.hooks = append(hs.hooks[:i], hs.hooks[i+1:]...)
			return
		}
	}
}

func (hs *hooks) Get() []hook {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	return hs.hooks
}
