package core_test

import (
	"testing"

	"github.com/janitorjeff/jeff-bot/core"
)

func eqSlices(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestSplit(t *testing.T) {
	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum lacinia quam turpis, sit amet ultrices odio scelerisque at. Sed nisl felis, efficitur ac eros vel, interdum viverra dui. Nam ut lectus mauris. Ut ac sem a ipsum mollis finibus. Praesent venenatis lorem vel urna tincidunt, ut interdum lacus efficitur. Nunc tellus nunc, euismod in quam eu, aliquam tincidunt nisl. Donec nec elementum lacus, eu porttitor turpis. Aenean rutrum, tortor et placerat pulvinar, quam libero auctor libero, vel malesuada massa felis at mauris. Interdum et malesuada fames ac ante ipsum primis in faucibus. Etiam nec augue quam."

	tests := []struct {
		text   string
		lenCnt func(string) int
		lenLim int
		res    []string
	}{
		{"Lorem", func(s string) int { return len(s) }, 2, []string{
			"Lo",
			"re",
			"m",
		}},
		// Res isn't {"Lorem", "ipsum"} because whitespace is also split
		// and included in the strings, and at the end any trailing whitespace
		// is trimmed. So in this case it would be {"Lorem", " ipsu", "m"}
		// which then gets trimmed to the expected result.
		{"Lorem ipsum", func(s string) int { return len(s) }, 5, []string{
			"Lorem",
			"ipsu",
			"m",
		}},
		{lorem, func(s string) int { return len(s) }, 50, []string{
			"Lorem ipsum dolor sit amet, consectetur",
			"adipiscing elit. Vestibulum lacinia quam turpis,",
			"sit amet ultrices odio scelerisque at. Sed nisl",
			"felis, efficitur ac eros vel, interdum viverra",
			"dui. Nam ut lectus mauris. Ut ac sem a ipsum",
			"mollis finibus. Praesent venenatis lorem vel urna",
			"tincidunt, ut interdum lacus efficitur. Nunc",
			"tellus nunc, euismod in quam eu, aliquam",
			"tincidunt nisl. Donec nec elementum lacus, eu",
			"porttitor turpis. Aenean rutrum, tortor et",
			"placerat pulvinar, quam libero auctor libero, vel",
			"malesuada massa felis at mauris. Interdum et",
			"malesuada fames ac ante ipsum primis in faucibus.",
			"Etiam nec augue quam.",
		}},
	}

	for _, test := range tests {
		if split := core.Split(test.text, test.lenCnt, test.lenLim); !eqSlices(split, test.res) {
			t.Fatalf("failed to split text, got %#v, expected %#v", split, test.res)
		}
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"google.com", false},
		{"google.com/test", false},
		{"https://google.com", true},
		{"http://google.com", true},
		{"http:/google.com", false},
		{"https://?test=", false},
		{"abc://google.com", false},
	}

	for _, test := range tests {
		if valid := core.IsValidURL(test.url); valid != test.valid {
			t.Fatalf("expected %t for url '%s', got %t", test.valid, test.url, valid)
		}
	}
}

func TestClean(t *testing.T) {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	upercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := `!"#$%&\'()*+,-./:;<=>?@[\]^_{|}~` + "`"
	whitespace := " \t\n\r"

	s := lowercase + special + upercase + whitespace + digits

	hope := lowercase + upercase + digits
	if cleaned := core.Clean(s); cleaned != hope {
		t.Fatalf("expected '%s' got '%s'", hope, cleaned)
	}
}
