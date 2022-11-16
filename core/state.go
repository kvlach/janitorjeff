package core

import (
	"crypto/rand"
	"encoding/base64"

	"git.slowtyper.com/slowtyper/gosafe"
)

// OAuth state parameter
type States struct {
	gosafe.Slice[string]
}

func (s *States) Delete(token string) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < s.LenUnsafe(); i++ {
		if s.GetUnsafe(i) == token {
			s.DeleteUnstableUnsafe(i)
		}
	}
}

func (s *States) generate() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// you never know
	if s.InUnsafe(state) {
		return s.generate()
	}
	return state, nil
}

func (s *States) Generate() (string, error) {
	// Because generate() recursively calls itself until it generates a state
	// that does not already exist we need to wrap it in a function that
	// handles locking and unlocking, otherwise we'd get a mutex deadlock.
	s.Lock()
	defer s.Unlock()
	return s.generate()
}
