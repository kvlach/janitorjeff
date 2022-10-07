package command

import (
	"log"
	"os"
	"testing"

	"git.slowtyper.com/slowtyper/janitorjeff/internal/testkit"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/discord"
	"github.com/rs/zerolog"
)

var (
	testScope  int64
	testAuthor int64
)

const (
	testTrigger        = "!test"
	testResponse       = "test response"
	testResponseModify = "modified response"
)

func TestDBAdd(t *testing.T) {
	if err := dbAdd(testScope, testAuthor, testTrigger, testResponse); err != nil {
		t.Fatalf("failed to create command: %v", err)
	}

	resp, err := dbGetResponse(testScope, testTrigger)
	if err != nil {
		t.Fatal(err)
	}
	if resp != testResponse {
		t.Fatalf("expected response '%s' got '%s'", testResponse, resp)
	}
}

func TestDBList(t *testing.T) {
	triggers, err := dbList(testScope)
	if err != nil {
		t.Fatal(err)
	}
	if len(triggers) != 1 || triggers[0] != testTrigger {
		t.Fatalf("expected '%v' got '%v'", []string{testTrigger}, triggers)
	}
}

func TestDBModify(t *testing.T) {
	if err := dbModify(testScope, testAuthor, testTrigger, testResponseModify); err != nil {
		t.Fatal(err)
	}

	resp, err := dbGetResponse(testScope, testTrigger)
	if err != nil {
		t.Fatal(err)
	}
	if resp != testResponseModify {
		t.Fatalf("expected response '%s' got '%s'", testResponseModify, resp)
	}
}

func TestDBDelete(t *testing.T) {
	if err := dbDel(testScope, testAuthor, testTrigger); err != nil {
		t.Fatal(err)
	}

	triggers, err := dbList(testScope)
	if err != nil {
		t.Fatal(err)
	}
	if len(triggers) != 0 {
		t.Fatalf("expected empty list of triggers got '%v' instead", triggers)
	}
}

func TestDBHistory(t *testing.T) {
	history, err := dbHistory(testScope, testTrigger)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 ||
		history[0].response != testResponse ||
		history[0].deleter == 0 ||
		history[1].response != testResponseModify ||
		history[1].deleter == 0 {

		t.Fatalf("got unexpected history '%v'", history)
	}
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tdb := testkit.NewTestDB("test.db")

	if err := tdb.Schema(dbShema); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	msg := testkit.NewTestMessage().DiscordRandom()

	var err error

	testScope, err = msg.Scope()
	if err != nil {
		log.Fatalln(err)
	}

	testAuthor, err = msg.Scope(discord.Author)
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()
	tdb.Delete()
	os.Exit(code)
}
