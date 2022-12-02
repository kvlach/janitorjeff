package audio

import (
	"encoding/json"
	"errors"
	"net/url"
	"os/exec"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/youtube"
	"github.com/janitorjeff/jeff-bot/core"

	"github.com/janitorjeff/gosafe"
)

var (
	ErrNotPaused        = errors.New("Not paused.")
	ErrNotPlaying       = errors.New("Not playing anything.")
	ErrSiteNotSupported = errors.New("This website is not supported.")
)

type Item struct {
	URL   string `json:"webpage_url"`
	Title string `json:"title"`
}

type Playing struct {
	State *core.State
	Queue *gosafe.Slice[Item]
}

var playing = gosafe.Map[int64, *Playing]{}

func stream(s core.Speaker, p *Playing, place int64) {
	for {
		if p.Queue.Len() == 0 {
			playing.Delete(place)
			return
		}

		switch p.State.Get() {
		case core.Play:
			// Audio only format might not exist in which case we grab the
			// whole thing and let ffmpeg extract the audio
			ytdl := exec.Command("yt-dlp", "-f", "bestaudio/best", "-o", "-", p.Queue.Get(0).URL)
			core.PipeThroughFFmpeg(s, ytdl, p.State)
			p.Queue.DeleteStable(0)
		case core.Stop:
			// Stop state means that the skip command was executed and so we
			// set the state to Play in order for the next item in the queue to
			// start
			p.State.Set(core.Play)
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

func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func Play(args []string, sp core.Speaker, place int64) (Item, error, error) {
	var item Item
	var err error

	if IsValidURL(args[0]) {
		item, err = GetInfo(args[0])
		if err != nil {
			return item, ErrSiteNotSupported, nil
		}
	} else {
		vid, usrErr, err := youtube.SearchVideo(strings.Join(args, " "))
		if usrErr != nil || err != nil {
			return item, usrErr, err
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

	state := &core.State{}
	state.Set(core.Play)

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

func Pause(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	if p.State.Get() != core.Play {
		return ErrNotPlaying
	}
	p.State.Set(core.Pause)
	return nil
}

func Resume(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	if p.State.Get() != core.Pause {
		return ErrNotPaused
	}
	p.State.Set(core.Play)
	return nil
}

func Skip(place int64) error {
	p, ok := playing.Get(place)
	if !ok {
		return ErrNotPlaying
	}
	p.State.Set(core.Stop)
	return nil
}

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
