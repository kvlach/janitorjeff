package twitch

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/utils"

	twitchIRC "github.com/gempir/go-twitch-irc/v2"
	"github.com/rs/zerolog/log"
)

type IRC struct {
	client  *twitchIRC.Client
	message *twitchIRC.PrivateMessage

	Helix *Helix
}

var twitchIrcClient *twitchIRC.Client

func (irc *IRC) Admin() bool {
	return false
}

func (irc *IRC) Parse() (*core.Message, error) {
	author := &core.Author{
		ID:          irc.message.User.ID,
		Name:        irc.message.User.Name,
		DisplayName: irc.message.User.DisplayName,
		Mention:     fmt.Sprintf("@%s", irc.message.User.DisplayName),
	}

	channel := &core.Channel{
		ID:   irc.message.RoomID,
		Name: irc.message.Channel,
	}

	msg := &core.Message{
		ID:      irc.message.ID,
		Type:    core.Twitch,
		Raw:     irc.message.Message,
		IsDM:    false,
		Author:  author,
		Channel: channel,
		Client:  irc,
	}

	// Ignore error since accessToken might not exist
	accessToken, _ := twitchChannelGetAccessToken(channel.ID)

	var err error
	// TODO: If no user access token, use app access token
	irc.Helix, err = HelixInit(core.Globals.Twitch.ClientID, accessToken)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (irc *IRC) checkID(id string) error {
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// not even a number so no point in asking twitch if it's valid
		return fmt.Errorf("id '%s' is not valid", id)
	}

	// try to get the id's corresponding username, if it fails then that means
	// that the id is not valid
	_, err := irc.Helix.GetUserName(id)
	return err
}

func (irc *IRC) getID(s string) (string, error) {
	// expected inputs are either a username, a mention (@username) or the
	// id itself

	if err := irc.checkID(s); err == nil {
		return s, nil
	}
	s = strings.TrimPrefix(s, "@")

	// try to get the corresponding id from the username, if it exists then
	// it will fetch and return with no error, if not then it will fail
	// and return an error
	return irc.Helix.GetUserID(s)
}

// Place and Person refer to the same thing on twitch
func (irc *IRC) PersonID(s, _ string) (string, error) {
	return irc.getID(s)
}

// Place and Person refer to the same thing on twitch
func (irc *IRC) PlaceID(s string) (string, error) {
	return irc.getID(s)
}

func (irc *IRC) PersonScope(id string) (int64, error) {
	return twitchChannelAddChannel(id, irc.message, irc.Helix)
}

func (irc *IRC) PlaceScope(id string) (int64, error) {
	return twitchChannelAddChannel(id, irc.message, irc.Helix)
}

func (irc *IRC) Write(msg any, _ error) (*core.Message, error) {
	var text string
	switch t := msg.(type) {
	case string:
		text = msg.(string)
	default:
		return nil, fmt.Errorf("Can't send twitch message of type %v", t)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	// This is how twitch's server seems to count the length, even though the
	// chat client on twitch's website doesn't follow this
	lenLim := 500
	lenCnt := utf8.RuneCountInString

	if lenLim > lenCnt(text) {
		irc.client.Say(irc.message.Channel, text)
		return nil, nil
	}

	parts := utils.Split(text, lenCnt, lenLim)
	for _, p := range parts {
		irc.client.Say(irc.message.Channel, p)
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

func onPrivateMessage(m twitchIRC.PrivateMessage) {
	irc := &IRC{client: twitchIrcClient, message: &m}
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

func IRCInit(stop chan struct{}, nick string, oauth string, channels []string) {
	twitchIrcClient = twitchIRC.NewClient(nick, oauth)

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

	<-stop

	log.Debug().Msg("closing twitch irc")
	if err = twitchIrcClient.Disconnect(); err != nil {
		log.Debug().Err(err).Msg("failed to close twitch irc connection")
	} else {
		log.Debug().Msg("closed twitch irc")
	}
}
