package twitch

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

type states struct {
	lock   sync.RWMutex
	tokens []string
}

func (s *states) Add(token string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.tokens = append(s.tokens, token)
}

func (s *states) In(token string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, t := range s.tokens {
		if t == token {
			return true
		}
	}
	return false
}

func (s *states) Delete(token string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for i, t := range s.tokens {
		if t == token {
			// delete the token by replacing it with the last one since the
			// order doesn't matter
			s.tokens[i] = s.tokens[len(s.tokens)-1]
			s.tokens = s.tokens[:len(s.tokens)-1]
		}
	}
}

var States = &states{}

func GetState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// you never know
	if States.In(state) {
		return GetState()
	}
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

		states, ok := q["state"]
		if !ok || len(states) == 0 {
			fail(w, r)
			return
		}
		state := states[0]

		codes, ok := q["code"]
		if !ok || len(codes) == 0 {
			fail(w, r)
			return
		}
		code := codes[0]

		c, err := helix.NewClient(&helix.Options{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			RedirectURI:  "http://" + core.Globals.Host + callback,
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

		if !States.In(state) {
			log.Debug().Str("state", state).Msg("got unexpected state")
			fail(w, r)
			return
		}

		States.Delete(state)

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
