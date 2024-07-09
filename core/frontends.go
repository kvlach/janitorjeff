package core

import (
	"fmt"
	"strings"
	"sync"
)

var Frontends Frontenders

type FrontendType int

var UrrUnknownFrontend = UrrNew("Can't recognize the provided frontend.")

type Frontender interface {
	// Type returns the frontend type ID.
	Type() FrontendType

	// Name returns frontend's name formatted in all lowercase.
	Name() string

	// Init is responsible for starting up any frontend-specific services and
	// connecting to frontend. When it receives the stop signal, then it should
	// disconnect from everything.
	Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{})

	// CreateMessage returns a Message object based on the given arguments.
	// Used to send messages that are not direct replies, e.g. reminders.
	CreateMessage(person, place int64, msgID string) (*EventMessage, error)

	// Usage returns the passed usage formatted appropriately for the frontend.
	Usage(usage string) any

	// PlaceExact returns the exact scope of the specified ID.
	PlaceExact(id string) (place int64, err error)

	// PlaceLogical returns the logical scope of the specified ID.
	PlaceLogical(id string) (place int64, err error)
}

type Frontenders []Frontender

// CreateMessage returns a Message object based on the given arguments. It
// detects what the frontend is based on the place. Used to send messages that
// are not direct replies, e.g. reminders.
func (fs Frontenders) CreateMessage(person, place int64, msgID string) (*EventMessage, error) {
	frontendType, err := DB.ScopeFrontend(place)
	if err != nil {
		return nil, err
	}

	for _, f := range fs {
		if f.Type() == FrontendType(frontendType) {
			return f.CreateMessage(person, place, msgID)
		}
	}

	return nil, fmt.Errorf("frontend type %d couldn't be matched", frontendType)
}

// Match returns the Frontender corresponding to lowercase fname.
// Returns UrrUnknownFrontend if nothing is matched.
func (fs Frontenders) Match(fname string) (Frontender, Urr) {
	fname = strings.ToLower(fname)
	for _, f := range fs {
		if f.Name() == fname {
			return f, nil
		}
	}
	return nil, UrrUnknownFrontend
}
