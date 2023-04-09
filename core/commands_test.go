package core_test

import (
	"testing"

	"github.com/janitorjeff/jeff-bot/commands/custom-command"
	"github.com/janitorjeff/jeff-bot/commands/help"
	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/internal/testkit"
)

// Using these commands as they don't require complicated permissions checks
// which potentially make calls to the respective frontend APIs
var cmds = core.CommandsStatic{
	help.Normal,
	help.Advanced,
	nick.Normal,
	nick.Advanced,
}

func TestFormat(t *testing.T) {
	hope := "#command delete <trigger>"
	if fmt := core.Format(custom_command.AdvancedDelete, "#"); fmt != hope {
		t.Fatalf("failed to format command, expected '%s', got '%s", hope, fmt)
	}
}

func TestCommandUsage(t *testing.T) {
	cmd := core.Command{
		CommandStatic: custom_command.AdvancedDelete,
		CommandRuntime: core.CommandRuntime{
			Path:   []string{"cmd", "rm"},
			Args:   []string{"test-cmd"},
			Prefix: "$",
		},
	}

	hope := "$cmd rm <trigger>"
	if usage := cmd.Usage(); usage != hope {
		t.Fatalf("failed to get usage, expected '%s', got '%s'", hope, usage)
	}
}

func TestMatch(t *testing.T) {
	msg := &testkit.NewTestMessage().DiscordRandom().Message
	match, index, err := cmds.Match(core.Advanced, msg, []string{"nick", "set", "test-cmd"})

	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	hope := nick.AdvancedSet
	if match != hope {
		t.Fatalf("didn't match command, expected '%s', got '%s'", core.Format(hope, ""), core.Format(match, ""))
	}

	if index != 1 {
		t.Fatalf("incorrect index, expected 1, got %d", index)
	}

}

func TestCommandsStaticUsage(t *testing.T) {
	hope := "(help | help | nick | nick)"
	if usage := cmds.Usage(); usage != hope {
		t.Fatalf("incorrect usage, expected '%s', got '%s'", hope, usage)
	}
}

func TestCommandsStaticOptionalUsage(t *testing.T) {
	hope := "[help | help | nick | nick]"
	if usage := cmds.UsageOptional(); usage != hope {
		t.Fatalf("incorrect usage, expected '%s', got '%s'", hope, usage)
	}
}
