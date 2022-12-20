package youtube

import (
	"errors"
)

var ErrVidNotFound = errors.New("No video was found.")

// SearchVideo will automatically create a client and return the most relevant
// completed video (not livestreams or premiers). Returns ErrVidNotFound if no
// video was found.
func SearchVideo(query string) (Video, error, error) {
	client, err := New()
	if err != nil {
		return Video{}, nil, err
	}

	vids, err := client.SearchVideos(query, 1)
	if err != nil {
		return Video{}, nil, err
	}

	if len(vids) == 0 {
		return Video{}, ErrVidNotFound, nil
	}

	return vids[0], nil, nil
}
