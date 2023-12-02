package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix/v2"
	"github.com/redis/go-redis/v9"
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

var ctx = context.Background()

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

		timestamp := c.GetHeader("Twitch-Eventsub-Message-Timestamp")
		when, err := time.Parse("2006-01-02T15:04:05.999999999Z", timestamp)
		if err != nil {
			log.Error().
				Err(err).
				Str("timestamp", timestamp).
				Msg("failed to parse timestamp")
			return
		}
		if time.Now().Sub(when) > 10*time.Minute {
			log.Error().Msg("message older than 10 minutes, skipping")
			return
		}

		id := c.GetHeader("Twitch-Eventsub-Message-Id")
		rdbKey := "frontend_twitch_eventsub_" + id

		if err := core.RDB.Get(ctx, rdbKey).Err(); err != redis.Nil {
			log.Debug().
				Str("id", id).
				Msg("message id has already been processed, skipping")
			return
		}

		log.Debug().Str("id", id).Msg("caching eventsub message id")
		if err := core.RDB.Set(ctx, rdbKey, nil, 10*time.Minute).Err(); err != nil {
			log.Error().Err(err).Str("id", id).Msg("failed to cache event id")
			return
		}

		switch t := vals.Subscription.Type; t {
		case "stream.online":
			var onlineEvent helix.EventSubStreamOnlineEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&onlineEvent)
			log.Debug().Msgf("got online webhook for channel: %s\n", onlineEvent.BroadcasterUserName)

			a, err := NewAuthor(
				onlineEvent.BroadcasterUserID,
				onlineEvent.BroadcasterUserLogin,
				onlineEvent.BroadcasterUserName,
				onlineEvent.BroadcasterUserID,
				nil,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("online", onlineEvent).
					Msg("failed to parse online event's author")
				return
			}

			h, err := NewHere(
				onlineEvent.BroadcasterUserID,
				onlineEvent.BroadcasterUserLogin,
				a,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("online", onlineEvent).
					Msg("failed to parse online event's here")
				return
			}

			on := &core.StreamOnline{
				When:     onlineEvent.StartedAt.Time,
				Here:     h,
				Frontend: Frontend,
			}

			core.EventStreamOnline <- on

		case "stream.offline":
			var offlineEvent helix.EventSubStreamOfflineEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&offlineEvent)
			log.Debug().Msgf("got offline webhook for channel: %s\n", offlineEvent.BroadcasterUserName)

			a, err := NewAuthor(
				offlineEvent.BroadcasterUserID,
				offlineEvent.BroadcasterUserLogin,
				offlineEvent.BroadcasterUserName,
				offlineEvent.BroadcasterUserID,
				nil,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("offline", offlineEvent).
					Msg("failed to parse offline event's author")
				return
			}

			h, err := NewHere(
				offlineEvent.BroadcasterUserID,
				offlineEvent.BroadcasterUserLogin,
				a,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("offline", offlineEvent).
					Msg("failed to parse offline event's here")
				return
			}

			off := &core.StreamOffline{
				When:     time.Now().UTC(),
				Here:     h,
				Frontend: Frontend,
			}
			core.EventStreamOffline <- off

		case "channel.channel_points_custom_reward_redemption.add":
			var redeem helix.EventSubChannelPointsCustomRewardRedemptionEvent
			err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeem)
			log.Debug().
				Str("broadcaster", redeem.BroadcasterUserName).
				Str("redeemer", redeem.UserName).
				Msg("got channel redeem event")

			a, err := NewAuthor(
				redeem.UserID,
				redeem.UserLogin,
				redeem.UserName,
				redeem.BroadcasterUserID,
				nil,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("redeem", redeem).
					Msg("failed to parse redeem's author")
				return
			}

			h, err := NewHere(
				redeem.BroadcasterUserID,
				redeem.BroadcasterUserName,
				a,
			)
			if err != nil {
				log.Error().
					Err(err).
					Interface("redeem", redeem).
					Msg("failed to parse redeem's here")
				return
			}

			r := &core.RedeemClaim{
				ID:       redeem.Reward.ID,
				Input:    redeem.UserInput,
				When:     redeem.RedeemedAt.Time,
				Author:   a,
				Here:     h,
				Frontend: Frontend,
			}

			core.EventRedeemClaim <- r

		default:
			log.Debug().Msgf("unhandled event type '%s'", t)
		}

		c.String(http.StatusOK, "ok")
	})
}
