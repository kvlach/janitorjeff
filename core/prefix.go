package core

import (
	"fmt"
)

var Prefixes = prefixes{}

// Prefixes:
//
// Some very basic prefix support has to exist in the core in order to be able
// to find scope specific prefixes when parsing the message. Only thing that is
// supported is creating the table and getting a list of prefixes for the
// current scope. The rest (adding, deleting, etc.) is handled externally by a
// command.
//
// The idea of having the ability to have an arbitrary number of prefixes
// originated from the fact that I wanted to not have to use a prefix when
// DMing the bot, as it is pointless to do so in that case. So after building
// support for more than 1 prefix it felt unecessary to limit it to just DMs,
// as that would be an artificial limit on what the bot is already supports and
// not really a necessity.
//
// There are 3 main prefix types:
//  - normal
//  - advanced
//  - admin
// These are used to call normal, advanced and admin commands respectively. A
// prefix has to be unique accross all types (in a specific scope), for example
// the prefix `!` can't be used for both normal and advanced commands.

type Prefix struct {
	Type   CommandType
	Prefix string
}

type prefixes struct {
	admin  []Prefix
	others []Prefix
}

func (ps *prefixes) Add(t CommandType, p string) {
	switch t {
	case Admin:
		ps.admin = append(ps.admin, Prefix{t, p})
	case Normal, Advanced:
		ps.others = append(ps.others, Prefix{t, p})
	default:
		panic(fmt.Sprintf("Unexpected prefix type: %d", t))
	}
}

func (ps prefixes) Admin() []Prefix {
	// Return a copy of the original slice since the returned slice might be
	// modified which would affect the default prefixes.
	admin := make([]Prefix, len(ps.admin))
	copy(admin, ps.admin)
	return admin
}

func (ps prefixes) Others() []Prefix {
	// Return a copy of the original slice since the returned slice might be
	// modified which would affect the default prefixes.
	others := make([]Prefix, len(ps.others))
	copy(others, ps.others)
	return others
}
