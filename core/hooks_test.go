package core_test

import (
	"testing"

	"github.com/janitorjeff/jeff-bot/core"
)

func TestHooks(t *testing.T) {
	id := core.Hooks.Register(func(m *core.Message) {
		return
	})

	hooks := core.Hooks.Get()
	if len(hooks) != 1 {
		t.Fatalf("expected length of hooks to be 1, got %d", len(hooks))
	}
	if hooks[0].ID != id {
		t.Fatalf("expected hook id to be %d, got %d", id, hooks[0].ID)
	}

	core.Hooks.Delete(id)

	hooks = core.Hooks.Get()
	if len(hooks) != 0 {
		t.Fatalf("expected hook slice to be empty, got %v", hooks)
	}
}
