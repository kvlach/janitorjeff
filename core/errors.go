package core

import (
	"errors"
)

// Urr represents a user-induced error (as opposed to an unexpected internal
// error.) This means that if such an error is returned, the user should, in
// some way, be informed of what their mistake was and not just receive a
// message along the lines of "something went wrong"
type Urr error

func UrrNew(text string) Urr {
	return errors.New(text)
}

var (
	UrrMissingArgs = UrrNew("not enough arguments provided")
	UrrSilence     = UrrNew("if this error is returned don't send any message")
)
