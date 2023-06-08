package youtube

import (
	"context"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	service *youtube.Service
}

// New returns a new YouTube client using the key set in core.YouTubeKey.
func New() (*Client, error) {
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(core.YouTubeKey))
	if err != nil {
		return nil, err
	}
	return &Client{service}, nil
}
