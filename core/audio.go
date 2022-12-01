package core

import (
	"github.com/janitorjeff/gosafe"
)

const (
	Play int = iota
	Pause
	Stop
)

// A select with multiple ready cases chooses one pseudo-randomly. So if the
// goroutine is "slow" to check those channels, you might send a value on both
// pause and resume (assuming they are buffered) so receiving from both channels
// could be ready, and resume could be chosen first, and in a later iteration
// the pause when the goroutine should not be paused anymore.
//
// Source: https://stackoverflow.com/a/60490371
type State struct {
	gosafe.Value[int]
}
