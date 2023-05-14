package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/commands"
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		fName := runtime.FuncForPC(pc).Name()
		return fmt.Sprintf("%s", path.Base(fName))
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Stack().Caller().Logger()
	}
}

func init() {
	log.Debug().Msg("starting event loop")
	go core.EventLoop()
}

func readVar(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		log.Fatal().Msgf("no $%s given", name)
	}
	log.Debug().Str(name, v).Msg("read env variable")
	return v
}

func connect(stop chan struct{}, wgStop *sync.WaitGroup) {
	// TODO: Handle inability to connect to a specific platform more gracefully,
	// in case something is down

	wgInit := new(sync.WaitGroup)
	wgInit.Add(len(frontends.Frontends))
	wgStop.Add(len(frontends.Frontends))

	twitch.Frontend.Nick = "JanitorJeff"
	twitch.Frontend.OAuth = readVar("TWITCH_OAUTH")
	twitch.Frontend.Channels = strings.Split(readVar("TWITCH_CHANNELS"), ",")

	discord.Frontend.Token = readVar("DISCORD_TOKEN")

	for _, f := range frontends.Frontends {
		go f.Init(wgInit, wgStop, stop)
	}

	wgInit.Wait()
}

func main() {
	log.Debug().Msg("opening db")
	dbConn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		readVar("POSTGRES_USER"),
		readVar("POSTGRES_PASSWORD"),
		readVar("POSTGRES_DB"),
		readVar("POSTGRES_HOST"),
		readVar("POSTGRES_PORT"),
		readVar("POSTGRES_SSLMODE"),
	)
	db, err := core.Open("postgres", dbConn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open DB")
	}
	defer func(db *core.SQLDB) {
		log.Debug().Msg("closing db")
		if err := db.Close(); err != nil {
			log.Warn().Err(err).Msg("couldn't close database")
		}
	}(db)

	log.Debug().Msg("connecting to redis")
	core.RDB = redis.NewClient(&redis.Options{
		Addr: readVar("REDIS_ADDR"),
	})

	core.Frontends = frontends.Frontends
	core.Commands = &commands.Commands
	core.DB = db
	core.Port = readVar("PORT")
	core.VirtualHost = readVar("VIRTUAL_HOST")
	core.YouTubeKey = readVar("YOUTUBE")
	core.TikTokSessionID = readVar("TIKTOK_SESSION_ID")
	core.OpenAIKey = readVar("OPENAI_KEY")

	minGodIntervalSeconds, err := strconv.Atoi(readVar("MIN_GOD_INTERVAL_SECONDS"))
	if err != nil {
		panic("invalid MIN_GOD_INTERVAL_SECONDS value, expected a number")
	}
	core.MinGodInterval = time.Duration(minGodIntervalSeconds) * time.Second

	core.Prefixes.Add(core.Admin, "##")
	core.Prefixes.Add(core.Normal, "!")
	core.Prefixes.Add(core.Advanced, "$")

	discord.Admins = strings.Split(readVar("DISCORD_ADMINS"), ",")
	twitch.Admins = strings.Split(readVar("TWITCH_ADMINS"), ",")
	twitch.ClientID = readVar("TWITCH_CLIENT_ID")
	twitch.ClientSecret = readVar("TWITCH_CLIENT_SECRET")

	stop := make(chan struct{})
	wgStop := new(sync.WaitGroup)
	connect(stop, wgStop)

	commands.Init()

	if err = core.Gin.SetTrustedProxies([]string{core.VirtualHost}); err != nil {
		log.Warn().Err(err).Msg("failed to set trusted proxies for gin")
	}
	go func() {
		if err := core.Gin.Run(":" + core.Port); err != nil {
			log.Fatal().Err(err).Msg("gin failed while running")
		}
	}()

	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	close(stop)
	wgStop.Wait()
}
