package twitch

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/kvlach/janitorjeff/core"

	tirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"github.com/rs/zerolog/log"
)

const Type = 1 << 1

var (
	Admins       []string
	ClientID     string
	ClientSecret string
)

type frontend struct {
	Nick     string
	OAuth    string
	Channels []string
}

var Frontend = &frontend{}

type Twitch struct {
	Author core.Personifier
	Here   core.Placer
	Client *tirc.Client
}

func NewTwitch(author core.Personifier, here core.Placer, client *tirc.Client) *Twitch {
	return &Twitch{
		Author: author,
		Here:   here,
		Client: client,
	}
}

func NewMessage(m tirc.PrivateMessage) (*core.Message, error) {
	a, err := NewAuthor(
		m.User.ID,
		m.User.Name,
		m.User.DisplayName,
		m.RoomID,
		m.User.Badges,
	)
	if err != nil {
		return nil, err
	}
	h, err := NewHere(m.RoomID, m.Channel, a)
	if err != nil {
		return nil, err
	}
	return &core.Message{
		ID:       m.ID,
		Raw:      m.Message,
		Frontend: Frontend,
		Author:   a,
		Here:     h,
		Client:   NewTwitch(a, h, twitchIrcClient),
		Speaker:  Speaker{},
	}, nil
}

func (f *frontend) Type() core.FrontendType {
	return Type
}

func (f *frontend) Name() string {
	return "twitch"
}

func (f *frontend) Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{}) {
	twitchIrcClient = tirc.NewClient(f.Nick, f.OAuth)

	twitchIrcClient.OnPrivateMessage(func(pm tirc.PrivateMessage) {
		m, err := NewMessage(pm)
		if err != nil {
			log.Error().
				Err(err).
				Interface("msg", pm).
				Msg("failed to parse private message")
		}
		core.EventMessage <- m
	})

	twitchIrcClient.Join(f.Channels...)

	log.Debug().Msg("connecting to twitch irc")
	var err error
	// FIXME: Race condition??? (err doesn't have time to be set)
	go func() {
		err = twitchIrcClient.Connect()
	}()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to twitch irc")
	} else {
		log.Debug().Msg("connected to twitch irc")
	}

	if err := generateAppAccessToken(); err != nil {
		panic(err)
	}

	wgInit.Done()
	<-stop

	log.Debug().Msg("closing twitch irc")
	if err = twitchIrcClient.Disconnect(); err != nil {
		log.Debug().Err(err).Msg("failed to close twitch irc connection")
	} else {
		log.Debug().Msg("closed twitch irc")
	}
	wgStop.Done()
}

func (f *frontend) CreateMessage(person, place int64, _ string) (*core.Message, error) {
	personID, err := dbGetChannel(person)
	if err != nil {
		return nil, err
	}
	placeID, err := dbGetChannel(place)
	if err != nil {
		return nil, err
	}

	pm := tirc.PrivateMessage{
		RoomID: placeID,
		User: tirc.User{
			ID: personID,
		},
	}
	return NewMessage(pm)
}

func (f *frontend) Usage(usage string) any {
	return fmt.Sprintf("Usage: %s", usage)
}

func (f *frontend) PlaceExact(id string) (int64, error) {
	return dbAddChannel(id)
}

func (f *frontend) PlaceLogical(id string) (int64, error) {
	return f.PlaceExact(id)
}

func (f *frontend) Helix() (*Helix, error) {
	hx, err := helix.NewClient(&helix.Options{
		ClientID:       ClientID,
		AppAccessToken: appAccessToken.Get(),
	})
	if err != nil {
		return nil, err
	}
	return &Helix{hx}, nil
}

var twitchIrcClient *tirc.Client

func (t *Twitch) Helix() (*Helix, error) {
	roomID, err := t.Here.IDExact()
	if err != nil {
		return nil, err
	}
	return NewHelix(roomID)
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (t *Twitch) Parse() (*core.Message, error) {
	panic("TODO: THIS WILL BE DELETED")
}

func (t *Twitch) checkID(id string) error {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// not even a number so no point in asking twitch if it's valid
		return fmt.Errorf("id '%s' is not valid", id)
	}

	hx, err := t.Helix()
	if err != nil {
		return err
	}

	// try to get the id's corresponding user, if it fails then that means that
	// the id is not valid
	_, err = hx.GetUser(id)

	return err
}

func (t *Twitch) getID(s string) (string, error) {
	// expected inputs are either a username, a mention (@username) or the
	// id itself

	if err := t.checkID(s); err == nil {
		return s, nil
	}
	s = strings.TrimPrefix(s, "@")

	hx, err := t.Helix()
	if err != nil {
		return "", err
	}

	// try to get the corresponding id from the username, if it exists then
	// it will fetch and return with no error, if not then it will fail
	// and return an error
	return hx.GetUserID(s)
}

// Place and Person refer to the same thing on twitch
func (t *Twitch) PersonID(s, _ string) (string, error) {
	return t.getID(s)
}

// Place and Person refer to the same thing on twitch
func (t *Twitch) PlaceID(s string) (string, error) {
	return t.getID(s)
}

func (t *Twitch) Person(id string) (int64, error) {
	return dbAddChannel(id)
}

func (t *Twitch) send(msg any, mention string) (*core.Message, error) {
	var text string
	switch t := msg.(type) {
	case string:
		text = msg.(string)
	default:
		return nil, fmt.Errorf("Can't send twitch message of type %v", t)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	// This is how twitch's server seems to count the length, even though the
	// chat client on twitch's website doesn't follow this. Subtract the
	// mention's length since it is added to every message sent.
	lenLim := 500 - len(mention)
	lenCnt := utf8.RuneCountInString

	ch, err := t.Here.Name()
	if err != nil {
		return nil, err
	}

	if lenLim > lenCnt(text) {
		t.Client.Say(ch, mention+text)
		return nil, nil
	}

	parts := core.Split(text, lenCnt, lenLim)
	for _, p := range parts {
		t.Client.Say(ch, mention+p)
	}

	return nil, nil
}

func (t *Twitch) Send(msg any, _ core.Urr) (*core.Message, error) {
	return t.send(msg, "")
}

func (t *Twitch) Ping(msg any, _ core.Urr) (*core.Message, error) {
	dn, err := t.Author.DisplayName()
	if err != nil {
		return nil, err
	}
	mention := fmt.Sprintf("@%s -> ", dn)
	return t.send(msg, mention)
}

func (t *Twitch) Write(msg any, urr core.Urr) (*core.Message, error) {
	return t.Ping(msg, urr)
}

func (t *Twitch) Natural(msg any, _ core.Urr) (*core.Message, error) {
	var mention string
	var err error
	// need this to only happen 30% of the time
	if num := core.Rand().Intn(10); num < 3 {
		mention, err = t.Author.DisplayName()
		if err != nil {
			return nil, err
		}
	}
	return t.send(msg, mention)
}

func (t *Twitch) QuoteCommand(cmd string) string {
	return "'" + cmd + "'"
}
