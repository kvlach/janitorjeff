package core

import (
	"sync"
)

type Frontender interface {
	Init(wgInit, wgStop *sync.WaitGroup, stop chan struct{})
}
