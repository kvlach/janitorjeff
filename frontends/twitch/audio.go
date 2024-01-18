package twitch

import (
	"io"

	"github.com/kvlach/janitorjeff/core"
)

type Speaker struct{}

func (s Speaker) Enabled() bool {
	return false
}

func (s Speaker) FrameRate() int {
	return 0
}

func (s Speaker) Channels() int {
	return 0
}

func (s Speaker) Join() error {
	return nil
}

func (s Speaker) Leave() error {
	return nil
}

func (s Speaker) Say(io.Reader, <-chan core.AudioState) error {
	return nil
}

func (s Speaker) AuthorDeafened() (bool, error) {
	return false, nil
}

func (s Speaker) AuthorConnected() (bool, error) {
	return false, nil
}
