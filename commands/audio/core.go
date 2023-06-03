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

type Playing struct {
	State *core.AudioState
	Queue *gosafe.Slice[Item]
}

var playing = gosafe.Map[int64, *Playing]{}

func stream(sp core.AudioSpeaker, p *Playing, place int64) {
	for {
		if p.Queue.Len() == 0 {
			playing.Delete(place)
			return
		}

		switch p.State.Get() {
		case core.AudioPlay, core.AudioLoop:
			// Audio only format might not exist in which case we grab the
			// whole thing and let ffmpeg extract the audio
			ytdl := exec.Command("yt-dlp", "-f", "bestaudio/best", "-o", "-", p.Queue.Get(0).URL)
			if err := core.AudioProcessCommand(sp, ytdl, p.State); err != nil {
				log.Error().Err(err).Msg("failed to stream audio")
				return
			}
			if p.State.Get() != core.AudioLoop {
				p.Queue.DeleteStable(0)
			}
		case core.AudioStop:
			// Stop state means that the skip command was executed and so we
			// set the state to Play in order for the next item in the queue to
			// start
			p.State.Set(core.AudioPlay)
		}
	}
}

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
		p.Queue.Append(item)
		return item, nil, nil
	}

	err = sp.Join()
	if err != nil {
		return item, nil, err
	}

	state := &core.AudioState{}
	state.Set(core.AudioPlay)

	queue := &gosafe.Slice[Item]{}
	queue.Append(item)

	p := &Playing{
		State: state,
		Queue: queue,
	}
	go stream(sp, p, place)
	playing.Set(place, p)

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
	if p.State.Get() != core.AudioPlay {
		return ErrNotPlaying
	}
	p.State.Set(core.AudioPause)
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
	if p.State.Get() != core.AudioPause {
		return ErrNotPaused
	}
	p.State.Set(core.AudioPlay)
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
	p.State.Set(core.AudioStop)
	return nil
}

// LoopOn will turn on looping by setting the state to Loop for the specified
// place. Returns an ErrNotPlaying if the queue is empty.
func LoopOn(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	p.State.Set(core.AudioLoop)
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
	if p.State.Get() != core.AudioLoop {
		return ErrNotLooping
	}
	p.State.Set(core.AudioPlay)
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

	p.Queue.RLock()
	defer p.Queue.RUnlock()

	var items []Item
	for i := 0; i < p.Queue.LenUnsafe(); i++ {
		items = append(items, p.Queue.GetUnsafe(i))
	}

	return items, nil
}
