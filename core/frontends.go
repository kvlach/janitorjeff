package core

import (
	"fmt"
	"sync"
)

var Frontends Frontenders

type FrontendType int

type Frontender interface {
	// Type returns the frontend type ID.
	Type() FrontendType

	// Init is responsible for starting up any frontend specific services and
	// connecting to frontend. When it receives the stop signal then it should
	// disconnect from everything.
	Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{})

	// CreateMessage returns a Message object based on the given arguments.
	// Used to send messages that are not direct replies, e.g. reminders.
	CreateMessage(person, place int64, msgID string) (*Message, error)
}

type Frontenders []Frontender

// CreateMessage returns a Message object based on the given arguments. It
// detects what the frontend is based on the place. Used to send messages that
// are not direct replies, e.g. reminders.
func (fs Frontenders) CreateMessage(person, place int64, msgID string) (*Message, error) {
	frontendType, err := DB.ScopeFrontend(place)
	if err != nil {
		return nil, err
	}

	for _, f := range Frontends {
		if f.Type() == FrontendType(frontendType) {
			return f.CreateMessage(person, place, msgID)
		}
	}

	return nil, fmt.Errorf("no frontend matched")
}
