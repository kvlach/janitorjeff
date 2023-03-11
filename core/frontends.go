package core

import (
	"sync"
)

type FrontendType int

type Frontender interface {
	// Type returns the frontend type ID.
	Type() FrontendType

	// Init is responsible for starting up any frontend specific services and
	// connecting to frontend. When it receives the stop signal then it should
	// disconnect from everything.
	Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{})
}
