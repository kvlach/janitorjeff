package core

import (
	"fmt"
	"sort"

	"github.com/rs/zerolog/log"
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

// Returns the given place's prefixes and also whether or not they were taken
// from the database (if not then that means the default ones were used).
func PlacePrefixes(place int64) ([]Prefix, bool, error) {
	// Initially the empty prefix was added if a message came from a DM, so
	// that normal commands could be run without using any prefix. This was
	// dropped because it added some unecessary complexity since we couldn't
	// always trivially know whether a place was a DM or not.

	prefixes, err := DB.PrefixList(place)
	if err != nil {
		return nil, false, err
	}

	log.Debug().
		Int64("place", place).
		Interface("prefixes", prefixes).
		Msg("place specific prefixes")

	inDB := true
	if len(prefixes) == 0 {
		inDB = false
		prefixes = Prefixes.Others()
		log.Debug().Msg("no place specific prefixes, using defaults")
	}

	// The admin prefixes remain constant across places and can only be
	// modified through the config. This means that they are never saved in the
	// database and so we just append them to the list every time. This doesn't
	// affect the `inDB` return value.
	prefixes = append(prefixes, Prefixes.Admin()...)

	// We order by the length of the prefix in order to avoid matching the
	// wrong prefix. For example, if the prefixes `!` and `!!` both exist in
	// the same place and `!` is placed first in the list of prefixes then it
	// will always get matched. So even if the user uses `!!`, the command will
	// be parsed as having the `!` prefix and will fail to match (since it will
	// try to match an invalid command, `!test` for example, instead of
	// trimming both '!' first).
	//
	// The prefixes *must* be sorted as a whole and cannot be split into
	// seperate categories (for example having 3 different slices for the 3
	// different types of prefixes) as each prefix is unique across all
	// categories which means that the same reasoning that was described above
	// still applies.
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i].Prefix) > len(prefixes[j].Prefix)
	})

	log.Debug().
		Int64("place", place).
		Interface("prefixes", prefixes).
		Msg("got prefixes")

	return prefixes, inDB, nil
}
