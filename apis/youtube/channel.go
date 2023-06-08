package youtube

type Channel struct {
	ID    string
	Title string
}

func (ch Channel) URL() string {
	return "https://www.youtube.com/channel/" + ch.ID
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
