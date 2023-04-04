package twitch

import (
	"net/http"

	"github.com/janitorjeff/jeff-bot/core"

	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

var states = &core.States{}

func NewState() (string, error) {
	return states.New()
}

func init() {
	callback := "/twitch/callback"

	fail := func(c *gin.Context) {
		c.String(http.StatusUnauthorized, "Something went wrong...")
	}

	core.Gin.GET(callback, func(c *gin.Context) {
		q := c.Request.URL.Query()

		statesQuery, ok := q["state"]
		if !ok || len(statesQuery) == 0 {
			fail(c)
			return
		}
		state := statesQuery[0]

		if !states.In(state) {
			log.Debug().Str("state", state).Msg("got unexpected state")
			fail(c)
			return
		}
		states.Delete(state)

		codes, ok := q["code"]
		if !ok || len(codes) == 0 {
			fail(c)
			return
		}
		code := codes[0]

		client, err := helix.NewClient(&helix.Options{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			RedirectURI:  "https://" + core.VirtualHost + callback,
		})
		if err != nil {
			log.Debug().Err(err).Send()
			fail(c)
			return
		}

		resp, err := client.RequestUserAccessToken(code)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(c)
			return
		}

		accessToken := resp.Data.AccessToken
		refreshToken := resp.Data.RefreshToken

		isValid, res, err := client.ValidateToken(accessToken)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(c)
			return
		}
		if !isValid {
			log.Debug().Msg("token invalid")
			fail(c)
			return
		}

		userID := res.Data.UserID

		scope, err := dbAddChannelSimple(userID, res.Data.Login)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error")
			return
		}

		err = dbSetUserAccessToken(scope, accessToken, refreshToken)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(c)
			return
		}

		c.String(http.StatusOK, "Success!!!")
	})
}
