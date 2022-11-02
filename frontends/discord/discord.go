package discord

import (
	"sync"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const Type = 1 << 0

func Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{}, token string) {
	if err := dbInit(); err != nil {
		log.Fatal().Err(err).Msg("failed to init discord db schema")
	}

	d, err := dg.New("Bot " + token)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create discord client")
	}

	d.AddHandler(messageCreate)
	d.AddHandler(messageEdit)
	d.AddHandler(messageDelete)
	d.AddHandler(interactionCreate)

	// TODO: Specify only needed intents
	d.Identify.Intents = dg.MakeIntent(dg.IntentsAll)

	d.State.MaxMessageCount = 100

	log.Debug().Msg("connecting to discord")
	if err = d.Open(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to discord")
	} else {
		log.Debug().Msg("connected to discord")
		core.Globals.Discord.Client = d
	}

	wgInit.Done()
	<-stop

	log.Debug().Msg("closing discord")
	if err = d.Close(); err != nil {
		log.Debug().Err(err).Msg("failed to close discord connection")
	} else {
		log.Debug().Msg("closed discord")
	}
	wgStop.Done()
}
