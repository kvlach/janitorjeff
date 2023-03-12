package discord

import (
	"sync"

	"github.com/janitorjeff/jeff-bot/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const Type = 1 << 0

var (
	Session *dg.Session
	Admins  []string

	EmbedColor    = 0xAD88E0
	EmbedErrColor = 0xB14D4D
)

type frontend struct {
	Token string
}

var Frontend = &frontend{}

func (f *frontend) Type() core.FrontendType {
	return Type
}

func (f *frontend) Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{}) {
	if err := dbInit(); err != nil {
		log.Fatal().Err(err).Msg("failed to init discord db schema")
	}

	d, err := dg.New("Bot " + f.Token)
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
		Session = d
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

func (f *frontend) CreateMessage(author, channel int64, msgID string) (*core.Message, error) {
	channelID, err := core.DB.ScopeID(channel)
	if err != nil {
		return nil, err
	}

	guild, err := dbGetGuildFromChannel(channel)
	if err != nil {
		return nil, err
	}

	guildID, err := core.DB.ScopeID(guild)
	if err != nil {
		return nil, err
	}

	authorID, err := core.DB.ScopeID(author)
	if err != nil {
		return nil, err
	}

	// check if message id still exists (could have been deleted for example)
	if _, err := Session.ChannelMessage(channelID, msgID); err != nil {
		msgID = ""
	}

	d := &Message{
		Message: &dg.Message{
			ID:        msgID,
			ChannelID: channelID,
			GuildID:   guildID,
			Author: &dg.User{
				ID:  authorID,
				Bot: false,
			},
		},
	}

	return d.Parse()
}
