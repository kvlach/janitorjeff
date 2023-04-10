package core

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/janitorjeff/gosafe"
	"github.com/rivo/uniseg"
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

// Create a new state and add it to the list of states for 1 minute after which
// it is removed.
func (s *States) New() (string, error) {
	state, err := s.Generate()
	if err != nil {
		return "", err
	}

	s.Append(state)
	go func() {
		time.Sleep(1 * time.Minute)
		s.Delete(state)
	}()

	return state, nil
}

func OnlyOneBitSet(n int) bool {
	// https://stackoverflow.com/a/28303898
	return n&(n-1) == 0
}

// Monitor incoming messages until `check` is true or until timeout. If nothing
// is matched then the returned object will be nil.
func Await(timeout time.Duration, check func(*Message) bool) *Message {
	var m *Message

	found := make(chan bool)

	id := Hooks.Register(func(candidate *Message) {
		if check(candidate) {
			m = candidate
			found <- true
		}
	})

	select {
	case <-found:
		break
	case <-time.After(timeout):
		break
	}

	Hooks.Delete(id)
	return m
}

func splitGraphemeClusters(text string, lenCnt func(string) int, lenLim int, parts []string) []string {
	parts = append(parts, "")
	gr := uniseg.NewGraphemes(text)

	for gr.Next() {
		partLen := lenCnt(parts[len(parts)-1])
		runes := string(gr.Runes())

		if lenLim > partLen+len(runes) {
			parts[len(parts)-1] += runes
		} else {
			parts = append(parts, runes)
		}
	}

	return parts
}

// Splits a message into submessages. Tries to not split words unless it
// absolutely has to in which case it splits based on grapheme clusters.
func Split(text string, lenCnt func(string) int, lenLim int) []string {
	parts := []string{""}
	r := regexp.MustCompile(`[^\s]+|\s+`)

	// TODO: Also try to fit whole lines instead of just words.

	for _, word := range r.FindAllString(text, -1) {
		wordLen := lenCnt(word)
		partLen := lenCnt(parts[len(parts)-1])

		if lenLim > partLen+wordLen {
			parts[len(parts)-1] += word
		} else if lenLim > wordLen {
			parts = append(parts, word)
		} else {
			parts = splitGraphemeClusters(word, lenCnt, lenLim, parts)
		}
	}

	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}

	return parts
}

// IsValidURL returns true if the provided string is a valid URL with an http
// or https scheme and a host.
func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

// Clean returns a string with every character except the ones in the a-z, A-Z
// and 0-9 ranges stripped from the passed string. Assumes ASCII.
func Clean(s string) string {
	var b strings.Builder

	// the simplified string will, at most, have len(s) number of bytes
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]

		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
			b.WriteByte(c)
		}
	}

	return b.String()
}
