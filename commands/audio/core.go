package audio

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"git.sr.ht/~slowtyper/janitorjeff/commands/youtube"
	"git.sr.ht/~slowtyper/janitorjeff/core"

	"git.sr.ht/~slowtyper/gosafe"
	"github.com/rs/zerolog/log"
)

//goland:noinspection GoErrorStringFormat
var (
	ErrNotPaused        = errors.New("Not paused.")
	ErrNotPlaying       = errors.New("Not playing anything.")
	ErrNotLooping       = errors.New("Not looping.")
	ErrSiteNotSupported = errors.New("This website is not supported.")
)

type Item struct {
	URL   string `json:"webpage_url"`
	Title string `json:"title"`
}

var playing = gosafe.Map[int64, *core.AudioPlayer[Item]]{}

func GetInfo(url string) (Item, error) {
	ytdl := exec.Command("yt-dlp", "-j", url)
	stdout, err := ytdl.StdoutPipe()
	if err != nil {
		return Item{}, err
	}

	if err := ytdl.Start(); err != nil {
		return Item{}, err
	}

	var info Item
	if err := json.NewDecoder(stdout).Decode(&info); err != nil {
		return Item{}, err
	}

	return info, ytdl.Wait()
}

// Play will:
//   - Check if the first argument is a url, if yes, then will try to stream it.
//     If not then assumes that it is a search and will query youtube to find
//     the corresponding video.
//   - Join the voice channel if necessary.
//   - Adds the item in the queue, if no queue exists creates one and begins
//     item playback.
//
// Returns ErrSiteNotSupported if the provided URL is a website that is not
// supported. Also passes any potential user errors that were generated by
// youtube.SearchVideo if a video search was performed.
func Play(args []string, sp core.AudioSpeaker, place int64) (Item, error, error) {
	var item Item
	var err error

	if core.IsValidURL(args[0]) {
		item, err = GetInfo(args[0])
		if err != nil {
			return item, ErrSiteNotSupported, nil
		}
	} else {
		vid, urr, err := youtube.SearchVideo(strings.Join(args, " "))
		if urr != nil || err != nil {
			return item, urr, err
		}
		item = Item{
			URL:   vid.URL(),
			Title: vid.Title,
		}
	}

	if p, ok := playing.Get(place); ok {
		p.Append(item)
		return item, nil, nil
	}

	err = sp.Join()
	if err != nil {
		return item, nil, err
	}

	p := &core.AudioPlayer[Item]{}
	p.Append(item)
	p.HandlePlay(func(item Item, st <-chan core.AudioState) error {
		// Audio only format might not exist in which case we grab the
		// whole thing and let ffmpeg extract the audio
		ytdl := exec.Command("yt-dlp", "-f", "bestaudio/best", "-o", "-", item.URL)
		if err := core.AudioProcessCommand(sp, ytdl, st); err != nil {
			log.Error().Err(err).Msg("failed to stream audio")
			return err
		}
		return nil
	})

	playing.Set(place, p)

	go p.Start()

	return item, nil, nil
}

// Pause will pause by setting the state to Pause in the specified place.
// Returns ErrNotPlaying if the queue is empty or if the state is not set to
// Play.
func Pause(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	if p.Current() != core.AudioPlay {
		return ErrNotPlaying
	}
	p.Pause()
	return nil
}

// Resume will resume playback by setting the state to Play in the specified
// place. Returns an ErrNotPlaying if the queue is empty. Returns ErrNotPaused
// if the state is not set to Pause.
func Resume(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	if p.Current() != core.AudioPause {
		return ErrNotPaused
	}
	p.Play()
	return nil
}

// Skip will skip the currenly playing item by setting the state to Stop in
// the specified place, which will exit the currenly playing item allowing for
// the next one to be played. Returns ErrNotPlaying if the queue is empty.
func Skip(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	p.Stop()
	return nil
}

// LoopOn will turn on looping by setting the state to Loop for the specified
// place. Returns an ErrNotPlaying if the queue is empty.
func LoopOn(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	p.LoopAll()
	return nil
}

// LoopOff will turn looping off by setting the state to Play for the specified
// place. Returns ErrNotPlaying if the queue is empty. Returns ErrNotLooping if
// the current state is not Loop.
func LoopOff(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	if p.Current() != core.AudioLoopAll {
		return ErrNotLooping
	}
	p.Play()
	return nil
}

// Queue returns the list of items that are currenly in the queue. The first
// item is the one that is currenly active. Returns ErrNotPlaying if the queue
// is empty.
func Queue(place int64) ([]Item, error) {
	p, ok := playing.Get(place)
	if !ok {
		return nil, ErrNotPlaying
	}
	return p.Queue(), nil
}
