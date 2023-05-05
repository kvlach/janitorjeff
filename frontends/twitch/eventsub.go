package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

type eventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

const (
	CallbackEventSub = "/twitch/eventsub"
	secret           = "secretword"
)

func init() {
	core.Gin.POST(CallbackEventSub, func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Debug().Err(err).Msg("failed to read request body")
			return
		}
		defer c.Request.Body.Close()

		if !helix.VerifyEventSubNotification(secret, c.Request.Header, string(body)) {
			log.Warn().Msg("no valid signature on subscription")
			return
		} else {
			log.Debug().Msg("valid signature for subscription")
		}
		fmt.Println(string(body))

		var vals eventSubNotification
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals)
		if err != nil {
			log.Warn().Err(err).Msg("failed to decode json")
			return
		}
		// if there's a challenge in the request, respond with only the challenge to verify your eventsub.
		if vals.Challenge != "" {
			c.String(http.StatusOK, vals.Challenge)
			return
		}

		var onlineEvent helix.EventSubStreamOnlineEvent
		err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&onlineEvent)
		log.Debug().Msgf("got online webhook for channel: %s\n", onlineEvent.BroadcasterUserName)
		c.String(http.StatusOK, "ok")
	})
}

func sub() {
	client, err := helix.NewClient(&helix.Options{
		ClientID:       ClientID,
		AppAccessToken: appAccessToken.Get(),
	})
	if err != nil {
		panic(err)
	}

	resp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeStreamOnline,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: "709767271",
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://" + core.VirtualHost + CallbackEventSub,
			Secret:   secret,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", resp)
}
