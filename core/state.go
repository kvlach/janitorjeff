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

func (s *States) Generate() (string, error) {
	s.Lock()
	defer s.Unlock()

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// you never know
	if s.InUnsafe(state) {
		return s.Generate()
	}
	return state, nil
}
