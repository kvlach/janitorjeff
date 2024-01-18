package core

import (
	"github.com/kvlach/gosafe"
)

// Placer is the interface used to abstract the place where an event came from,
// e.g. channel, server, etc.
//
// Two type's of scopes exist for places, the exact and the logical. The logical
// is the area where things are generally expected to work. For example, if a
// user adds a custom command in a discord server, they would probably expect it
// to work in the entire server and not just in the specific channel that they
// added it in. If, on the other hand, someone adds a custom command in a discord
// DM, then no guild exists and thus the channel's scope would have to be used.
// On the other hand, the exact scope is, as its name suggests, the scope of the
// exact place the message came from and does not account for context, so using
// the previous discord server example, it would be the channel's scope where
// the message came from instead of the server's.
type Placer interface {
	// IDExact returns the exact ID, this should be a unique, static,
	// identifier in that frontend.
	IDExact() (id string, err error)

	// IDLogical returns the logical ID, this should be a unique, static,
	// identifier for the frontend.
	IDLogical() (id string, err error)

	// Name return's the channel's name.
	Name() (name string, err error)

	// ScopeExact returns the here's exact scope.
	// Returns the exact scope in Teleports if the author is present there.
	// See the interface's doc comment for more information on exact scopes.
	ScopeExact() (place int64, err error)

	// ScopeLogical returns the here's logical scope.
	// Returns the logical scope in Teleports if the author is present there.
	// See the interface's doc comment for more information on logical scopes.
	ScopeLogical() (place int64, err error)
}

type Place struct {
	Exact   int64
	Logical int64
}

// Teleports holds the bot-admin teleports defined by the teleport command.
// At most, only a handful of people are ever expected to be bot-admins,
// so using a map should suffice.
var Teleports = gosafe.Map[int64, Place]{}
