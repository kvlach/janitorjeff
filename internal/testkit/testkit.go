package testkit

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

type TestDB struct {
	name string
	db   *core.DB
}

func NewTestDB(name string) *TestDB {
	name = fmt.Sprintf("%d-%s", time.Now().UnixNano(), name)

	db, err := core.Open("sqlite3", name)
	if err != nil {
		log.Fatalf("failed to open DB: %v\n", err)
	}

	globals := &core.GlobalVars{
		DB: db,
	}
	core.GlobalsInit(globals)

	tdb := &TestDB{
		name: name,
		db:   db,
	}
	return tdb
}

func (tdb *TestDB) Delete() {
	err := os.Remove(tdb.name)
	if err != nil {
		log.Fatalf("failed to delete DB: %v\n", err)
	}
}

func (tdb *TestDB) Schema(schema string) error {
	return tdb.db.Init(schema)
}

type TestMessage struct {
	core.Message
}

var names = []string{
	"janitor",
	"jeff",
	"janitorjeff",
	"JanitorJeff",
}

func NewTestMessage() *TestMessage {
	return &TestMessage{core.Message{}}
}

func randomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func randomName() string {
	rand.Seed(time.Now().UnixNano())
	return names[rand.Intn(len(names))]
}

func (tm *TestMessage) DiscordRandom() *TestMessage {
	dgMsg := &dg.MessageCreate{
		Message: &dg.Message{
			ID:      randomID(),
			Content: "tmp",
			Author: &dg.User{
				ID:       randomID(),
				Username: randomName(),
			},
			ChannelID: randomID(),
			GuildID:   "123", // isDM
		},
	}

	msgTmp := discord.MessageCreate{Session: nil, Message: dgMsg}

	msg, err := msgTmp.Parse()
	if err != nil {
		log.Fatalln(err)
	}

	return &TestMessage{*msg}
}
