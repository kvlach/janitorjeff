package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix/v2"
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

		switch t := vals.Subscription.Type; t {
		case "stream.online":
			var onlineEvent helix.EventSubStreamOnlineEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&onlineEvent)
			log.Debug().Msgf("got online webhook for channel: %s\n", onlineEvent.BroadcasterUserName)

			h := Here{
				RoomID:   onlineEvent.BroadcasterUserID,
				RoomName: onlineEvent.BroadcasterUserLogin,
			}

			on := &core.StreamOnline{
				When: onlineEvent.StartedAt.Time,
				Here: h,
			}

			core.EventStreamOnline <- on

		case "stream.offline":
			var offlineEvent helix.EventSubStreamOfflineEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&offlineEvent)
			log.Debug().Msgf("got offline webhook for channel: %s\n", offlineEvent.BroadcasterUserName)

			h := Here{
				RoomID:   offlineEvent.BroadcasterUserID,
				RoomName: offlineEvent.BroadcasterUserLogin,
			}

			off := &core.StreamOffline{
				When: time.Now().UTC(),
				Here: h,
			}
			core.EventStreamOffline <- off

		case "channel.channel_points_custom_reward_redemption.add":
			var redeem helix.EventSubChannelPointsCustomRewardRedemptionEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeem)
			log.Debug().
				Str("broadcaster", redeem.BroadcasterUserName).
				Str("redeemer", redeem.UserName).
				Msg("got channel redeem event")

			a := Author{
				id:          redeem.UserID,
				username:    redeem.UserLogin,
				displayName: redeem.UserName,
				roomID:      redeem.BroadcasterUserID,
			}

			h := Here{
				RoomID:   redeem.BroadcasterUserID,
				RoomName: redeem.BroadcasterUserName,
			}

			r := &core.RedeemClaim{
				ID:     redeem.Reward.ID,
				When:   redeem.RedeemedAt.Time,
				Author: a,
				Here:   h,
			}

			core.EventRedeemClaim <- r

		default:
			log.Debug().Msgf("unhandled event type '%s'", t)
		}

		c.String(http.StatusOK, "ok")
	})
}
