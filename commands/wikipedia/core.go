package wikipedia

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var errNoResult = errors.New("no results found")

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

func search(query string) (page, error, error) {
	resp, err := http.Get(queryURL + url.QueryEscape(query))
	if err != nil {
		return page{}, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return page{}, nil, err
	}

	var r response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return page{}, nil, err
	}

	if len(r.Query.Pages) == 0 {
		return page{}, errNoResult, nil
	}

	return r.Query.Pages[0], nil, nil
}
