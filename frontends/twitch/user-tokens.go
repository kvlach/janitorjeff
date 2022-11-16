package twitch

import (
	"fmt"
	"net/http"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

var states = &core.States{}

func GetState() (string, error) {
	state, err := states.Generate()
	if err != nil {
		return "", err
	}

	states.Append(state)
	go func() {
		time.Sleep(1 * time.Minute)
		states.Delete(state)
	}()

	return state, nil
}

func init() {
	callback := "/twitch/callback"
	success := "/twitch/success"
	failure := "/twitch/failure"

	fail := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, failure, 400)
	}

	http.HandleFunc(callback, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		statesQuery, ok := q["state"]
		if !ok || len(statesQuery) == 0 {
			fail(w, r)
			return
		}
		state := statesQuery[0]

		if !states.In(state) {
			log.Debug().Str("state", state).Msg("got unexpected state")
			fail(w, r)
			return
		}
		states.Delete(state)

		codes, ok := q["code"]
		if !ok || len(codes) == 0 {
			fail(w, r)
			return
		}
		code := codes[0]

		c, err := helix.NewClient(&helix.Options{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			RedirectURI:  "http://" + core.Host + callback,
		})
		if err != nil {
			log.Debug().Err(err).Send()
			fail(w, r)
			return
		}

		resp, err := c.RequestUserAccessToken(code)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(w, r)
			return
		}

		accessToken := resp.Data.AccessToken
		refreshToken := resp.Data.RefreshToken

		isValid, res, err := c.ValidateToken(accessToken)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(w, r)
			return
		}
		if !isValid {
			log.Debug().Msg("token invalid")
			fail(w, r)
			return
		}

		userID := res.Data.UserID

		err = SetUserAccessToken(accessToken, refreshToken, userID)
		if err != nil {
			log.Debug().Err(err).Send()
			fail(w, r)
			return
		}

		http.Redirect(w, r, success, 301)
	})

	http.HandleFunc(success, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Success!!!")
	})

	http.HandleFunc(failure, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Something went wrong...")
	})
}
