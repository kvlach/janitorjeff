package twitch

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/utils"

	tirc "github.com/gempir/go-twitch-irc/v2"
	"github.com/rs/zerolog/log"
)

const Type = 1 << 1

type Twitch struct {
	client  *tirc.Client
	message *tirc.PrivateMessage

	Helix *Helix
}

func CreateClient(person, place int64) (*Twitch, error) {
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

	return t, nil
}

var twitchIrcClient *tirc.Client

func (t *Twitch) Admin() bool {
	return false
}

func (t *Twitch) Parse() (*core.Message, error) {
	user := &core.User{
		ID:          t.message.User.ID,
		Name:        t.message.User.Name,
		DisplayName: t.message.User.DisplayName,
		Mention:     fmt.Sprintf("@%s", t.message.User.DisplayName),
	}

	channel := &core.Channel{
		ID:   t.message.RoomID,
		Name: t.message.Channel,
	}

	msg := &core.Message{
		ID:      t.message.ID,
		Type:    Type,
		Raw:     t.message.Message,
		IsDM:    false,
		User:    user,
		Channel: channel,
		Client:  t,
	}

	// Ignore error since accessToken might not exist
	accessToken, _ := twitchChannelGetAccessToken(channel.ID)

	var err error
	// TODO: If no user access token, use app access token
	t.Helix, err = HelixInit(core.Globals.Twitch.ClientID, accessToken)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (t *Twitch) checkID(id string) error {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// not even a number so no point in asking twitch if it's valid
		return fmt.Errorf("id '%s' is not valid", id)
	}

	// try to get the id's corresponding username, if it fails then that means
	// that the id is not valid
	_, err := t.Helix.GetUserName(id)
	return err
}

func (t *Twitch) getID(s string) (string, error) {
	// expected inputs are either a username, a mention (@username) or the
	// id itself

	if err := t.checkID(s); err == nil {
		return s, nil
	}
	s = strings.TrimPrefix(s, "@")

	// try to get the corresponding id from the username, if it exists then
	// it will fetch and return with no error, if not then it will fail
	// and return an error
	return t.Helix.GetUserID(s)
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
	return twitchChannelAddChannel(id, t.message, t.Helix)
}

func (t *Twitch) PlaceExact(id string) (int64, error) {
	return twitchChannelAddChannel(id, t.message, t.Helix)
}

func (t *Twitch) PlaceLogical(id string) (int64, error) {
	return twitchChannelAddChannel(id, t.message, t.Helix)
}

func (t *Twitch) Usage(usage string) any {
	return fmt.Sprintf("Usage: %s", usage)
}

func (t *Twitch) Write(msg any, _ error) (*core.Message, error) {
	var text string
	switch t := msg.(type) {
	case string:
		text = msg.(string)
	default:
		return nil, fmt.Errorf("Can't send twitch message of type %v", t)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	mention := fmt.Sprintf("@%s -> ", t.message.User.DisplayName)

	// This is how twitch's server seems to count the length, even though the
	// chat client on twitch's website doesn't follow this. Subtract the
	// mention's length since it is added to every message sent.
	lenLim := 500 - len(mention)
	lenCnt := utf8.RuneCountInString

	if lenLim > lenCnt(text) {
		t.client.Say(t.message.Channel, fmt.Sprintf("%s%s", mention, text))
		return nil, nil
	}

	parts := utils.Split(text, lenCnt, lenLim)
	for _, p := range parts {
		t.client.Say(t.message.Channel, fmt.Sprintf("%s%s", mention, p))
	}

	return nil, nil
}

// func (tirc *TwitchIRC) Delete() error {
// 	_, err := tirc.Write(fmt.Sprintf("/delete %s", tirc.ID))
// 	return err
// }

// func (tirc *TwitchIRC) Edit(msg any) (*core.Message, error) {
// 	return nil, fmt.Errorf("editing not supported for twitch irc")
// }

func onPrivateMessage(m tirc.PrivateMessage) {
	irc := &Twitch{client: twitchIrcClient, message: &m}
	msg, err := irc.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	msg.Run()
}

// func IRCInit(nick string, oauth string, channels []string) *twitchIRC.Client {
// 	twitchIrcClient = twitchIRC.NewClient(nick, oauth)

// 	twitchIrcClient.OnPrivateMessage(onPrivateMessage)

// 	twitchIrcClient.Join(channels...)

// 	go twitchIrcClient.Connect()

// 	return twitchIrcClient
// }

func IRCInit(wgInit, wgStop *sync.WaitGroup, stop chan struct{}, nick string, oauth string, channels []string) {
	if err := dbInit(); err != nil {
		log.Fatal().Err(err).Msg("failed to init twitch db schema")
	}

	twitchIrcClient = tirc.NewClient(nick, oauth)

	twitchIrcClient.OnPrivateMessage(onPrivateMessage)

	twitchIrcClient.Join(channels...)

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
