package core

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

// OAuth state parameter

type States struct {
	lock   sync.Mutex
	tokens []string
}

func (s *States) Add(token string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.tokens = append(s.tokens, token)
}

func (s *States) in(token string) bool {
	for _, t := range s.tokens {
		if t == token {
			return true
		}
	}
	return false
}

func (s *States) In(token string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.in(token)
}

func (s *States) Delete(token string) {
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

func (s *States) Generate() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// you never know
	if s.in(state) {
		return s.Generate()
	}
	return state, nil
}
