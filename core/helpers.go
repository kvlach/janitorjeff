package core

import (
	"time"
)

func OnlyOneBitSet(n int) bool {
	// https://stackoverflow.com/a/28303898
	return n&(n-1) == 0
}

// Monitor incoming messages until `check` is true or until timeout. If nothing
// is matched then the returned object will be nil.
func Await(timeout time.Duration, check func(*Message) bool) *Message {
	var m *Message

	found := make(chan bool)

	id := Hooks.Register(func(candidate *Message) {
		if check(candidate) {
			m = candidate
			found <- true
		}
	})

	select {
	case <-found:
		break
	case <-time.After(timeout):
		break
	}

	Hooks.Delete(id)
	return m
}
