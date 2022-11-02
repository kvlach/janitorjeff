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
	"sync"
	"syscall"

	"git.slowtyper.com/slowtyper/janitorjeff/commands"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var myEnv map[string]string

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

func recurseCommands(cmd *core.CommandStatic) {
	if cmd.Init != nil {
		if err := cmd.Init(); err != nil {
			log.Fatal().Err(err).Msgf("failed to init command %v", cmd)
		}
	}

	if cmd.Children == nil {
		return
	}

	for _, child := range cmd.Children {
		child.Parent = cmd
		recurseCommands(child)
	}
}

// Setting the parents when declaring the object is not possible because that
// results in an inialization loop error (children reference the parent, so the
// parent can't reference the children). So instead we just recursively loop
// through all the children and set the parent. This also makes declaring the
// command objects a bit cleaner.
func commandsSetUp() {
	for _, cmd := range commands.Normal {
		recurseCommands(cmd)
	}

	for _, cmd := range commands.Advanced {
		recurseCommands(cmd)
	}

	for _, cmd := range commands.Admin {
		recurseCommands(cmd)
	}
}

func readVar(name string) string {
	v, ok := myEnv[name]
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
	wgInit.Add(2)
	wgStop.Add(2)

	// Twitch IRC
	twitchOauth := readVar("TWITCH_OAUTH")
	channels := strings.Split(readVar("TWITCH_CHANNELS"), ",")
	go twitch.IRCInit(wgInit, wgStop, stop, "JanitorJeff", twitchOauth, channels)

	// Discord
	discordToken := readVar("DISCORD_TOKEN")
	go discord.Init(wgInit, wgStop, stop, discordToken)

	wgInit.Wait()
}

func main() {
	// OTHER
	var err error
	myEnv, err = godotenv.Read("data/secrets.env")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read enviromental variables")
	}

	log.Debug().Msg("opening db")
	db, err := core.Open("sqlite3", "jeff.db")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open DB")
	}
	defer db.Close()
	defer log.Debug().Msg("closing db")

	globals := &core.GlobalVars{
		Commands: core.AllCommands{
			Normal:   commands.Normal,
			Advanced: commands.Advanced,
			Admin:    commands.Admin,
		},
		DB:   db,
		Host: readVar("HOST"),
		Prefixes: core.Prefixes{
			Admin: []core.Prefix{
				{Type: core.Admin, Prefix: "##"},
			},
			Others: []core.Prefix{
				{Type: core.Normal, Prefix: "!"},
				{Type: core.Advanced, Prefix: "$"},
			},
		},
		Discord: &core.DiscordVars{
			EmbedColor:    0xAD88E0,
			EmbedErrColor: 0xB14D4D,
			Admins: []string{
				"155662023743635456",
			},
		},

		Twitch: &core.TwitchVars{
			ClientID:     readVar("TWITCH_CLIENT_ID"),
			ClientSecret: readVar("TWITCH_CLIENT_SECRET"),
		},
	}

	core.GlobalsInit(globals)

	stop := make(chan struct{})
	wgStop := new(sync.WaitGroup)
	connect(stop, wgStop)

	// Requires globals to be set
	commandsSetUp()

	go http.ListenAndServe(globals.Host, nil)

	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	close(stop)
	wgStop.Wait()
}
