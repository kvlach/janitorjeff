package core

import (
	"errors"
)

var (
	ErrMissingArgs = errors.New("not enough arguments provided")
	ErrSilence     = errors.New("if this error is returned don't send any message")
	ErrNotMod      = errors.New("Mod permissions are required.")
	ErrNotAdmin    = errors.New("Admin permissions are required.")
)
