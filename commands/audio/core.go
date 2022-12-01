package audio

import (
	"encoding/json"
	"errors"
	"net/url"
	"os/exec"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/janitorjeff/gosafe"
)

var (
	ErrNotPaused  = errors.New("Not paused.")
	ErrNotPlaying = errors.New("Not playing anything.")
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

func stream(v *dg.VoiceConnection, p *Playing, place int64) {
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
			discord.PipeThroughFFmpeg(v, ytdl, p.State)
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
