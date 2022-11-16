package core

import (
	"errors"
)

var (
	ErrMissingArgs = errors.New("not enough arguments provided")
	ErrSilence     = errors.New("if this error is returned don't send any message")
)
