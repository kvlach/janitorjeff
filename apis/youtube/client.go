package youtube

import (
	"net/http"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	key     string
	service *youtube.Service
}

// New returns a new youtube client using the key set in core.YouTubeKey.
func New() (*Client, error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: core.YouTubeKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}

	return &Client{core.YouTubeKey, service}, nil
}
