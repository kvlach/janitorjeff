package custom_command_test

import (
	"log"
	"os"
	"testing"

	"github.com/janitorjeff/jeff-bot/commands/custom-command"
	"github.com/janitorjeff/jeff-bot/core"
	_ "github.com/janitorjeff/jeff-bot/internal/testing_init"
	"github.com/janitorjeff/jeff-bot/internal/testkit"

	"github.com/rs/zerolog"
)

var (
	place  int64
	person int64
)

const (
	trigger1     = "!test"
	trigger2     = "!command"
	response     = "test response"
	responseEdit = "modified response"
)

func addCmd(trigger string) (error, error) {
	return custom_command.Add(place, person, trigger, response)
}

func editCmd(trigger string) (error, error) {
	return custom_command.Edit(place, person, trigger, responseEdit)
}

func showResp(trigger string) (string, error) {
	return custom_command.Show(place, trigger)
}

func rmCmd(trigger string) (error, error) {
	return custom_command.Delete(place, person, trigger)
}

func TestAdd(t *testing.T) {
	core.Prefixes.Add(core.Advanced, "$")

	core.Commands = &core.CommandsStatic{
		custom_command.Advanced,
	}

	if usrErr, err := addCmd(trigger1); usrErr != nil || err != nil {
		t.Fatalf("failed to create command '%s': usrErr = %v, err = %v", trigger1, usrErr, err)
	}

	if usrErr, err := addCmd(trigger1); usrErr != custom_command.ErrTriggerExists || err != nil {
		t.Fatalf("expected TriggerExists user error, got: usrErr = %v, err = %v", usrErr, err)
	}

	if usrErr, err := addCmd("$command"); usrErr != custom_command.ErrBuiltinCommand || err != nil {
		t.Fatalf("expected BuiltinCommand user error, got: usrErr = %v, err = %v", usrErr, err)
	}

	if usrErr, err := addCmd(trigger2); usrErr != nil || err != nil {
		t.Fatalf("failed to add command '%s': usrErr = %v, err = %v", trigger2, usrErr, err)
	}

	if resp, err := showResp(trigger1); resp != response || err != nil {
		t.Fatalf("expected response '%s' got '%s', err = %v", response, resp, err)
	}

	if resp, err := showResp(trigger2); resp != response || err != nil {
		t.Fatalf("expected response '%s' got '%s', err = %v", response, resp, err)
	}
}

func TestList(t *testing.T) {
	cmds, err := custom_command.List(place)
	if err != nil {
		t.Fatalf("failed to get list of places: %v\n", err)
	}
	if len(cmds) != 2 {
		t.Fatalf("unexpected len of list of commands, expected 2 got %d, %v\n", len(cmds), cmds)
	}
	if cmds[0] != trigger1 {
		t.Fatalf("unexpected trigger, expected '%s' got '%s'", trigger1, cmds[0])
	}
	if cmds[1] != trigger2 {
		t.Fatalf("unexpected trigger, expected '%s' got '%s'", trigger2, cmds[1])
	}
}

func TestEdit(t *testing.T) {
	if usrErr, err := editCmd(trigger1); usrErr != nil || err != nil {
		t.Fatalf("failed to edit trigger '%s' to response '%s': usrErr = %v, err = %v", trigger1, responseEdit, usrErr, err)
	}

	if usrErr, err := editCmd("random"); usrErr != custom_command.ErrTriggerNotFound || err != nil {
		t.Fatalf("expected TriggerNotFound user error, got: usrErr = %v, err = %v", usrErr, err)
	}

	if resp, err := showResp(trigger1); resp != responseEdit || err != nil {
		t.Fatalf("expected response '%s' got '%s', err = %v", responseEdit, resp, err)
	}
}

func TestDelete(t *testing.T) {
	if usrErr, err := rmCmd(trigger1); usrErr != nil || err != nil {
		t.Fatalf("failed to delete trigger '%s', got: usrErr = %v, err = %v", trigger1, usrErr, err)
	}

	if usrErr, err := rmCmd(trigger1); usrErr != custom_command.ErrTriggerNotFound || err != nil {
		t.Fatalf("expected TriggerNotFound user error, got: usrErr = %v, err = %v", usrErr, err)
	}

	triggers, err := custom_command.List(place)
	if err != nil {
		t.Fatalf("failed to get list of places: %v\n", err)
	}
	if len(triggers) != 1 {
		t.Fatalf("unexpected len of list of commands, expected 1 got %d, %v\n", len(triggers), triggers)
	}
	if triggers[0] != trigger2 {
		t.Fatalf("unexpected trigger, expected '%s' got '%s'", trigger2, triggers[0])
	}
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tdb := testkit.NewTestDB()
	msg := testkit.NewTestMessage().DiscordRandom()

	var err error

	place, err = msg.Here.ScopeLogical()
	if err != nil {
		log.Fatalln(err)
	}

	person, err = msg.Author.Scope()
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()
	tdb.Delete()
	os.Exit(code)
}
