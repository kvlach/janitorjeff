package audio

import (
	"encoding/json"
	"net/url"
	"os/exec"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	"git.slowtyper.com/slowtyper/gosafe"
	dg "github.com/bwmarrin/discordgo"
)

type audio struct {
	state *core.State
	queue []string
}

var states = gosafe.Map[int64, *core.State]{}

func stream(v *dg.VoiceConnection, url string, s *core.State) {
	// Audio only format might not exist in which case we grab the whole thing
	// and let ffmpeg extract the audio
	ytdl := exec.Command("yt-dlp", "-f", "bestaudio/best", "-o", "-", url)
	discord.PipeThroughFFmpeg(v, ytdl, s)
}

type Info struct {
	Title string `json:"title"`
	URL   string `json:"webpage_url"`
}

func GetInfo(url string) (Info, error) {
	ytdl := exec.Command("yt-dlp", "-j", url)
	stdout, err := ytdl.StdoutPipe()
	if err != nil {
		return Info{}, err
	}

	if err := ytdl.Start(); err != nil {
		return Info{}, err
	}

	var info Info
	if err := json.NewDecoder(stdout).Decode(&info); err != nil {
		return Info{}, err
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
