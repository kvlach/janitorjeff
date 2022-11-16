package core

import (
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

func OnlyOneBitSet(n int) bool {
	// https://stackoverflow.com/a/28303898
	return n&(n-1) == 0
}

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

// Monitor incoming messages until `check` is true or until timeout.
func Await(timeout time.Duration, check func(*Message) bool) *Message {
	var m *Message

	timeoutchan := make(chan bool)

	id := Hooks.Register(func(candidate *Message) {
		if check(candidate) {
			m = candidate
			timeoutchan <- true
		}
	})

	select {
	case <-timeoutchan:
		break
	case <-time.After(timeout):
		break
	}

	Hooks.Delete(id)
	return m
}
