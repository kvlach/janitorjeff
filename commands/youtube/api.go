package youtube

import (
	"net/http"

	"github.com/janitorjeff/jeff-bot/core"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	key     string
	service *youtube.Service
}

type Video struct {
	id    string
	title string
}

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

func (c *Client) SearchVideos(query string, maxResults int64) ([]Video, error) {
	call := c.service.Search.List([]string{"id", "snippet"}).
		Q(query).
		MaxResults(maxResults)

	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	var videos []Video
	for _, item := range response.Items {
		if item.Id.Kind == "youtube#video" {
			vid := Video{item.Id.VideoId, item.Snippet.Title}
			videos = append(videos, vid)
		}
	}

	return videos, nil
}
