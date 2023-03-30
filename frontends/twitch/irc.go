package twitch

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/janitorjeff/jeff-bot/core"

	tirc "github.com/gempir/go-twitch-irc/v2"
	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

const Type = 1 << 1

var (
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
	client  *tirc.Client
	message *tirc.PrivateMessage
}

func onPrivateMessage(m tirc.PrivateMessage) {
	irc := &Twitch{client: twitchIrcClient, message: &m}
	msg, err := irc.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	msg.Run()
}

func (f *frontend) Type() core.FrontendType {
	return Type
}

func (f *frontend) Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{}) {
	twitchIrcClient = tirc.NewClient(f.Nick, f.OAuth)

	twitchIrcClient.OnPrivateMessage(onPrivateMessage)

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
	personID, personName, err := dbGetChannel(person)
	if err != nil {
		return nil, err
	}

	placeID, placeName, err := dbGetChannel(place)
	if err != nil {
		return nil, err
	}

	t := &Twitch{
		client: twitchIrcClient,
		message: &tirc.PrivateMessage{
			RoomID:  placeID,
			Channel: placeName,
			User: tirc.User{
				ID:          personID,
				Name:        personName,
				DisplayName: personName,
			},
		},
	}

	return t.Parse()
}

var twitchIrcClient *tirc.Client

func (t *Twitch) Helix() (*Helix, error) {
	h, err := helix.NewClient(&helix.Options{
		ClientID: ClientID,
	})
	if err != nil {
		return nil, err
	}

	userAccessToken, err := dbGetUserAccessToken(t.message.RoomID)
	if err == nil {
		h.SetUserAccessToken(userAccessToken)
	} else {
		h.SetAppAccessToken(appAccessToken.Get())
	}

	return &Helix{h}, nil
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (t *Twitch) Parse() (*core.Message, error) {
	author := Author{
		User: t.message.User,
	}

	here := Here{
		RoomID:   t.message.RoomID,
		RoomName: t.message.Channel,
	}

	msg := &core.Message{
		ID:       t.message.ID,
		Raw:      t.message.Message,
		Frontend: Frontend,
		Author:   author,
		Here:     here,
		Client:   t,
		Speaker:  t,
	}

	return msg, nil
}

func (t *Twitch) checkID(id string) error {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// not even a number so no point in asking twitch if it's valid
		return fmt.Errorf("id '%s' is not valid", id)
	}

	h, err := t.Helix()
	if err != nil {
		return err
	}

	// try to get the id's corresponding user, if it fails then that means that
	// the id is not valid
	_, err = h.GetUser(id)

	return err
}

func (t *Twitch) getID(s string) (string, error) {
	// expected inputs are either a username, a mention (@username) or the
	// id itself

	if err := t.checkID(s); err == nil {
		return s, nil
	}
	s = strings.TrimPrefix(s, "@")

	h, err := t.Helix()
	if err != nil {
		return "", err
	}

	// try to get the corresponding id from the username, if it exists then
	// it will fetch and return with no error, if not then it will fail
	// and return an error
	return h.GetUserID(s)
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
	h, err := t.Helix()
	if err != nil {
		return -1, err
	}
	return dbAddChannel(id, t.message.User, h)
}

func (t *Twitch) PlaceExact(id string) (int64, error) {
	h, err := t.Helix()
	if err != nil {
		return -1, err
	}
	return dbAddChannel(id, t.message.User, h)
}

func (t *Twitch) PlaceLogical(id string) (int64, error) {
	h, err := t.Helix()
	if err != nil {
		return -1, err
	}
	return dbAddChannel(id, t.message.User, h)
}

func (t *Twitch) Usage(usage string) any {
	return fmt.Sprintf("Usage: %s", usage)
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

	if lenLim > lenCnt(text) {
		t.client.Say(t.message.Channel, fmt.Sprintf("%s%s", mention, text))
		return nil, nil
	}

	parts := core.Split(text, lenCnt, lenLim)
	for _, p := range parts {
		t.client.Say(t.message.Channel, fmt.Sprintf("%s%s", mention, p))
	}

	return nil, nil
}

func (t *Twitch) Send(msg any, _ error) (*core.Message, error) {
	return t.send(msg, "")
}

func (t *Twitch) Ping(msg any, _ error) (*core.Message, error) {
	mention := fmt.Sprintf("@%s -> ", t.message.User.DisplayName)
	return t.send(msg, mention)
}

func (t *Twitch) Write(msg any, usrErr error) (*core.Message, error) {
	return t.Ping(msg, usrErr)
}

/////////////
//         //
// Speaker //
//         //
/////////////

func (t *Twitch) Enabled() bool {
	return false
}

func (t *Twitch) FrameRate() int {
	return 0
}

func (t *Twitch) Channels() int {
	return 0
}

func (t *Twitch) Join() error {
	return nil
}

func (t *Twitch) Say(io.Reader, *core.AudioState) error {
	return nil
}

func (t *Twitch) AuthorDeafened() (bool, error) {
	return false, nil
}

func (t *Twitch) AuthorConnected() (bool, error) {
	return false, nil
}
