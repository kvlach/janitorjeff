package youtube

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"

	"git.sr.ht/~slowtyper/janitorjeff/apis/youtube"
)

var (
	UrrVidNotFound     = core.UrrNew("No video was found.")
	UrrChannelNotFound = core.UrrNew("No channel was found.")
)

// SearchVideo returns the most relevant video (not livestream or premiere)
// based on the query.
// Returns UrrVidNotFound if no video was found.
func SearchVideo(query string) (youtube.Video, core.Urr, error) {
	client, err := youtube.New()
	if err != nil {
		return youtube.Video{}, nil, err
	}

	vids, err := client.SearchVideos(query, 1)
	if err != nil {
		return youtube.Video{}, nil, err
	}
	if len(vids) == 0 {
		return youtube.Video{}, UrrVidNotFound, nil
	}
	return vids[0], nil, nil
}

// SearchChannel returns the most relevant channel based on the query.
// Returns UrrChannelNotFound if no channel was found.
func SearchChannel(query string) (youtube.Channel, core.Urr, error) {
	client, err := youtube.New()
	if err != nil {
		return youtube.Channel{}, nil, err
	}

	chs, err := client.SearchChannels(query, 1)
	if err != nil {
		return youtube.Channel{}, nil, err
	}
	if len(chs) == 0 {
		return youtube.Channel{}, UrrChannelNotFound, nil
	}
	return chs[0], nil, nil
}
