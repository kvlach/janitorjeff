package testkit

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type TestDB struct {
	*core.SQLDB
}

func NewTestDB() *TestDB {
	conn, err := sql.Open("postgres", "user=postgres password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect: %v\n", err)
	}

	if _, err := conn.Exec("CREATE DATABASE test_db"); err != nil {
		log.Fatalf("failed to create test database: %v\n", err)
	}

	if _, err := conn.Exec("CREATE USER test_user WITH PASSWORD 'test_pass'"); err != nil {
		log.Fatalf("failed to set test database password: %v\n", err)
	}

	if _, err = conn.Exec("GRANT ALL PRIVILEGES ON DATABASE test_db TO test_user"); err != nil {
		log.Fatalf("failed to give prvileges to test_user: %v\n", err)
	}

	conn.Close()

	conn, err = sql.Open("postgres", "user=postgres password=postgres dbname=test_db sslmode=disable")
	if _, err := conn.Exec("GRANT ALL ON SCHEMA PUBLIC TO test_user;"); err != nil {
		log.Fatalf("failed to grant all on schema public: %v\n", err)
	}
	conn.Close()

	sqlDB, err := sql.Open("postgres", "user=test_user password=test_pass dbname=test_db sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to test_db: %v\n", err)
	}

	schema, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema file: %v\n", err)
	}

	db := &core.SQLDB{DB: sqlDB}
	if err := db.Init(string(schema)); err != nil {
		log.Fatalf("failed to init schema: %v\n", err)
	}
	core.RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	core.DB = db
	return &TestDB{db}
}

func (tdb *TestDB) Delete() {
	if err := tdb.DB.Close(); err != nil {
		log.Fatalf("failed to close testing DB: %v\n", err)
	}

	conn, err := sql.Open("postgres", "user=postgres password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect: %v\n", err)
	}
	defer conn.Close()

	if _, err := conn.Exec("DROP DATABASE test_db"); err != nil {
		log.Fatalf("failed to delete DB: %v\n", err)
	}

	if _, err := conn.Exec("DROP USER test_user"); err != nil {
		log.Fatalf("failed to delete DB: %v\n", err)
	}
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

	msgTmp := discord.MessageCreate{Message: dgMsg}

	msg, err := msgTmp.Parse()
	if err != nil {
		log.Fatalln(err)
	}

	return &TestMessage{*msg}
}
