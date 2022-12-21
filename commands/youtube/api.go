package youtube

import (
	"net/http"

	"github.com/janitorjeff/jeff-bot/core"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	ID    string
	Title string
}

func (v Video) URL() string {
	return "https://youtu.be/" + v.ID
}

type Channel struct {
	ID    string
	Title string
}

func (ch Channel) URL() string {
	return "https://www.youtube.com/channel/" + ch.ID
}

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

// SearchVideos returns a list of videos, of maxResults length, matching the
// query and ranked by relevance. Will only return videos, not livestreams or
// premiers.
func (c *Client) SearchVideos(query string, maxResults int64) ([]Video, error) {
	call := c.service.Search.List([]string{"id", "snippet"}).
		Q(query).
		Type("video").
		EventType("completed").
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

// SearchChannels returns a list of channels, of maxResults length, matching the
// query and ranked by relevance.
func (c *Client) SearchChannels(query string, maxResults int64) ([]Channel, error) {
	call := c.service.Search.List([]string{"id", "snippet"}).
		Q(query).
		Type("channel").
		MaxResults(maxResults)

	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	var chs []Channel
	for _, item := range response.Items {
		if item.Id.Kind == "youtube#channel" {
			ch := Channel{item.Id.ChannelId, item.Snippet.ChannelTitle}
			chs = append(chs, ch)
		}
	}

	return chs, nil
}
