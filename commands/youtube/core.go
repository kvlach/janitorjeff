package youtube

import (
	"errors"

	"git.sr.ht/~slowtyper/janitorjeff/apis/youtube"
)

//goland:noinspection GoErrorStringFormat
var (
	ErrVidNotFound     = errors.New("No video was found.")
	ErrChannelNotFound = errors.New("No channel was found.")
)

// SearchVideo will automatically create a client and return the most relevant
// completed video (not livestreams or premiers). Returns ErrVidNotFound if no
// video was found.
func SearchVideo(query string) (youtube.Video, error, error) {
	client, err := youtube.New()
	if err != nil {
		return youtube.Video{}, nil, err
	}

	vids, err := client.SearchVideos(query, 1)
	if err != nil {
		return youtube.Video{}, nil, err
	}

	if len(vids) == 0 {
		return youtube.Video{}, ErrVidNotFound, nil
	}

	return vids[0], nil, nil
}

// SearchChannel will automatically create a client and return the most relevant
// channel. Returns ErrChannelNotFound if no channel was found.
func SearchChannel(query string) (youtube.Channel, error, error) {
	client, err := youtube.New()
	if err != nil {
		return youtube.Channel{}, nil, err
	}

	chs, err := client.SearchChannels(query, 1)
	if err != nil {
		return youtube.Channel{}, nil, err
	}
	if len(chs) == 0 {
		return youtube.Channel{}, ErrChannelNotFound, nil
	}

	return chs[0], nil, nil
}
