package urban_dictionary

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	endpoint = "https://api.urbandictionary.com/v0"
	define   = "/define?term="
)

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

func cleanLinks(def definition) definition {
	re := regexp.MustCompile(`\[[ a-zA-Z0-9]+\]`)

	def.Definition = re.ReplaceAllStringFunc(def.Definition, func(m string) string {
		return m[1 : len(m)-1]
	})

	def.Example = re.ReplaceAllStringFunc(def.Example, func(m string) string {
		return m[1 : len(m)-1]
	})

	return def
}

func search(term string) (definition, error) {
	term = url.QueryEscape(term)

	resp, err := http.Get(endpoint + define + term)
	if err != nil {
		return definition{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return definition{}, err
	}

	var defs response
	err = json.Unmarshal(body, &defs)
	if err != nil {
		return definition{}, err
	}

	if len(defs.List) == 0 {
		return definition{}, errors.New("no results found")
	}

	return cleanLinks(defs.List[0]), nil
}
