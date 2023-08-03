package wikipedia

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"git.sr.ht/~slowtyper/janitorjeff/core"
)

var UrrNoResult = core.UrrNew("no results found")

const queryURL = "https://en.wikipedia.org/w/api.php?" +
	"action=query" +
	"&format=json" +
	"&redirects=1" +
	"&prop=info" +
	"&generator=prefixsearch" +
	"&formatversion=2" +
	"&inprop=url" +
	"&gpslimit=1" +
	"&gpssearch="

type page struct {
	Canonicalurl string `json:"canonicalurl"`
}

type response struct {
	Query struct {
		Pages []page `json:"pages"`
	} `json:"query"`
}

func Search(query string) (page, core.Urr, error) {
	resp, err := http.Get(queryURL + url.QueryEscape(query))
	if err != nil {
		return page{}, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return page{}, nil, err
	}

	var r response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return page{}, nil, err
	}

	if len(r.Query.Pages) == 0 {
		return page{}, UrrNoResult, nil
	}

	return r.Query.Pages[0], nil, nil
}
