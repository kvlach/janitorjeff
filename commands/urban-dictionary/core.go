package urban_dictionary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	dg "github.com/bwmarrin/discordgo"
)

const (
	base   = "https://api.urbandictionary.com/v0"
	define = "/define?term="
	random = "/random"
)

var errNoResults = errors.New("No results found.")

type definition struct {
	Definition  string    `json:"definition"`
	Permalink   string    `json:"permalink"`
	ThumbsUp    int       `json:"thumbs_up"`
	SoundUrls   []string  `json:"sound_urls"`
	Author      string    `json:"author"`
	Word        string    `json:"word"`
	Defid       int       `json:"defid"`
	CurrentVote string    `json:"current_vote"`
	WrittenOn   time.Time `json:"written_on"`
	Example     string    `json:"example"`
	ThumbsDown  int       `json:"thumbs_down"`
}

type response struct {
	List []definition `json:"list"`
}

// term links are wrapped between `[]`, this removes the brackets
func cleanLinks(s string) string {
	re := regexp.MustCompile(`\[[ \-a-zA-Z]+\]`)

	return re.ReplaceAllStringFunc(s, func(m string) string {
		return m[1 : len(m)-1]
	})
}

func read(u string) (definition, error, error) {
	resp, err := http.Get(u)
	if err != nil {
		return definition{}, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return definition{}, nil, err
	}

	var defs response
	err = json.Unmarshal(body, &defs)
	if err != nil {
		return definition{}, nil, err
	}

	if len(defs.List) == 0 {
		return definition{}, errNoResults, nil
	}

	def := defs.List[0]
	def.Definition = cleanLinks(def.Definition)
	def.Example = cleanLinks(def.Example)
	return def, nil, nil
}

func Search(term string) (definition, error, error) {
	return read(base + define + url.QueryEscape(term))
}

func Random() (definition, error, error) {
	return read(base + random)
}

func renderDiscord(def definition, usrErr error) *dg.MessageEmbed {
	if usrErr != nil {
		return &dg.MessageEmbed{Description: fmt.Sprint(usrErr)}
	}

	var example []*dg.MessageEmbedField
	if def.Example != "" {
		example = []*dg.MessageEmbedField{
			{
				Name:  "Example",
				Value: def.Example,
			},
		}
	}

	embed := &dg.MessageEmbed{
		Title:       "UrbanDictionary definition for " + def.Word,
		URL:         def.Permalink,
		Description: def.Definition,
		Fields:      example,
		Footer: &dg.MessageEmbedFooter{
			Text: fmt.Sprintf("Submitter: %s | Thumbs up: %d | Thumbs down: %d", def.Author, def.ThumbsUp, def.ThumbsDown),
		},
	}

	return embed
}

func renderText(def definition, usrErr error) string {
	if usrErr != nil {
		return fmt.Sprint(usrErr)
	}

	def.Definition = strings.ReplaceAll(def.Definition, "\n", " ")
	return fmt.Sprintf("%s: %s %s", def.Word, def.Definition, def.Permalink)
}
