package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/commands"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/discord"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/twitch"
	"git.slowtyper.com/slowtyper/janitorjeff/sqldb"

	// TODO: Remove dependency
	"github.com/joho/godotenv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		fName := runtime.FuncForPC(pc).Name()
		return fmt.Sprintf("%s:%d | %s", file, line, path.Base(fName))
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Stack().Caller().Logger()
	}
}

func setParents(cmd *core.CommandStatic) {
	if cmd.Children == nil {
		return
	}

	for _, child := range cmd.Children {
		child.Parent = cmd
		setParents(child)
	}
}

// Setting the parents when declaring the object is not possible because that
// results in an inialization loop error (children reference the parent, so the
// parent can't reference the children). So instead we just recursively loop
// through all the children and set the parent. This also makes declaring the
// command objects a bit cleaner.
func init() {
	for _, cmd := range commands.Commands {
		setParents(cmd)
	}
}

func main() {
	// OTHER
	myEnv, err := godotenv.Read("secrets.env")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read enviromental variables")
	}

	readVar := func(name string) string {
		v, ok := myEnv[name]
		if !ok {
			log.Fatal().Msgf("no $%s given", name)
		}
		log.Debug().Str(name, v).Msg("read env variable")
		return v
	}

	log.Debug().Msg("opening db")
	db, err := sqldb.Open("sqlite3", "jeff.db")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open DB")
	}
	defer db.Close()
	defer log.Debug().Msg("closing db")

	globals := &core.GlobalVars{
		Commands: commands.Commands,
		DB:       db,
		Host:     readVar("HOST"),
		Prefixes_: []string{
			"!",
		},

		Discord: &core.DiscordVars{
			EmbedColor: 0xAD88E0,
		},

		Twitch: &core.TwitchVars{
			ClientID:     readVar("TWITCH_CLIENT_ID"),
			ClientSecret: readVar("TWITCH_CLIENT_SECRET"),
		},
	}

	core.GlobalsInit(globals)

	// Requires globals to be set
	for _, cmd := range commands.Commands {
		if cmd.Init != nil {
			if err := cmd.Init(); err != nil {
				log.Fatal().Err(err).Msgf("failed to init command %v", cmd)
			}
		}

		if len(cmd.Children) > 0 {
			log.Debug().Msgf("%v", cmd.Children[0])
		}
	}

	go http.ListenAndServe(globals.Host, nil)

	stop := make(chan struct{})

	// TODO: Handle inability to connect to a specific platform more gracefully,
	// in case something is down

	// Twitch IRC
	twitchOauth := readVar("TWITCH_OAUTH")
	channels := strings.Split(readVar("TWITCH_CHANNELS"), ",")
	go twitch.IRCInit(stop, "JanitorJeff", twitchOauth, channels)

	// Discord
	discordToken := readVar("DISCORD_TOKEN")
	go discord.Init(stop, discordToken)

	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	close(stop)

	// Give time for everything to close gracefully
	time.Sleep(5 * time.Second)
}
