package core

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

var Prefixes = prefixes{}

// Prefixes:
//
// Some very basic prefix support has to exist in the core in order to be able
// to find scope-specific prefixes when parsing the message.
// The only thing that is supported is creating the table and getting a list of
// prefixes for the current scope.
// The rest (adding, deleting, etc.) is handled externally by a command.
//
// The idea of having the ability to have an arbitrary number of prefixes
// originated from the fact that I wanted to not have to use a prefix when
// DMing the bot, as it is pointless to do so in that case.
// So after building support for more than 1 prefix, it felt unnecessary to
// limit it to just DMs, as that would be an artificial limit on what the bot is
// already supports and not really a necessity.
//
// There are 3 main prefix types:
//  - normal
//  - advanced
//  - admin
// These are used to call normal, advanced and admin commands respectively.
//A
// prefix has to be unique across all types (in a specific scope), for example,
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

// PlacePrefixes returns the given place's prefixes, and also whether they were
// taken from the database (if not, then that means the default ones were used).
func PlacePrefixes(place int64) ([]Prefix, bool, error) {
	// Initially, the empty prefix was added if a message came from a DM, so
	// that normal commands could be run without using any prefix. This was
	// dropped because it added some unnecessary complexity since we couldn't
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
	// database, and so we just append them to the list every time. This doesn't
	// affect the `inDB` return value.
	prefixes = append(prefixes, Prefixes.Admin()...)

	// We order by the length of the prefix in order to avoid matching the
	// wrong prefix. For example, if the prefixes `!` and `!!` both exist in
	// the same place and `!` is placed first in the list of prefixes, then it
	// will always get matched. So even if the user uses `!!`, the command will
	// be parsed as having the `!` prefix and will fail to match (since it will
	// try to match an invalid command, `!test`, for example, instead of
	// trimming both '!' first).
	//
	// The prefixes *must* be sorted as a whole and cannot be split into
	// separate categories (for example, having 3 different slices for the 3
	// different types of prefixes) as each prefix is unique across all
	// categories, which means that the same reasoning that was described above
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

// Messenger is the abstraction layer for message events.
type Messenger interface {
	Parse() (*Message, error)

	// PlaceID returns the ID of the passed string. The returned ID must be
	// valid. Generally used for verifying an ID's validity and extracting IDs
	// from mentions.
	PlaceID(s string) (id string, err error)

	// PersonID returns the ID of the passed string. The returned ID must be
	// valid. Generally used for verifying an ID's validity and extracting IDs
	// from mentions.
	PersonID(s, placeID string) (id string, err error)

	// Person returns the target's scope.
	// If it doesn't exist, it will create it and add it to the database.
	Person(id string) (person int64, err error)

	// PlaceExact returns the exact scope of the specified ID.
	PlaceExact(id string) (place int64, err error)

	// PlaceLogical returns the logical scope of the specified ID.
	PlaceLogical(id string) (place int64, err error)

	// Send sends a message to the appropriate scope, resp could be nil
	// depending on the frontend.
	Send(msg any, urr Urr) (resp *Message, err error)

	// Ping works the same as Send except the user is also pinged.
	Ping(msg any, urr Urr) (resp *Message, err error)

	// Write either calls Send or Ping depending on the frontend. This is what
	// should be used in most cases.
	Write(msg any, urr Urr) (resp *Message, err error)

	// Natural will try to emulate a response as if an actual human had written
	// it. Often the bot uses markers to distinguish its responses (for example,
	// on Twitch it replies with the following format: @person -> <resp>), which
	// are not natural looking. To add to the effect, randomness may be used to
	// only sometimes mention the person.
	Natural(msg any, urr Urr) (resp *Message, err error)
}

type Message struct {
	ID       string
	Raw      string
	Frontend Frontender
	Author   Personifier
	Here     Placer
	Client   Messenger
	Speaker  AudioSpeaker
	Command  *Command
}

func (m *Message) Fields() []string {
	return strings.Fields(m.Raw)
}

// FieldsSpace splits text into fields that include all trailing whitespace. For
// example, "example of    text" will be split into ["example ", "of    ", "text"]
func (m *Message) FieldsSpace() []string {
	text := strings.TrimSpace(m.Raw)
	re := regexp.MustCompile(`\S+\s*`)
	fields := re.FindAllString(text, -1)

	log.Debug().
		Str("text", text).
		Strs("fields", fields).
		Msg("split text into fields that include whitespace")

	return fields
}

// RawArgs returns the arguments including the whitespace between them. Skips
// over first n args. Pass 0 to not skip any.
func (m *Message) RawArgs(n int) string {
	if 0 > n {
		panic("unexpected n")
	}

	fields := m.FieldsSpace()

	// Skip over the command + the given offset
	s := strings.Join(fields[len(m.Command.Path)+n:], "")

	log.Debug().
		Int("offset", n).
		Str("raw-args", s).
		Msg("extracted raw arguments")

	return s
}

// Sends a message.
func (m *Message) Write(msg any, urr error) (*Message, error) {
	return m.Client.Write(msg, urr)
}

// Prefixes returns the logical here's prefixes, and also whether they were
// taken from the database (if not, then that means the default ones were used).
func (m *Message) Prefixes() ([]Prefix, bool, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, false, err
	}
	return PlacePrefixes(here)
}

func hasPrefix(prefixes []Prefix, s string) (Prefix, bool) {
	//goland:noinspection SpellCheckingInspection
	for _, p := range prefixes {
		// Example:
		// !prefix add !prefix
		// !prefixprefix ls // works
		// !prefix ls // doesn't work
		//
		// This is because the rootCmdName "!prefix" in the third command gets
		// matched as the prefix "!prefix" and not the prefix "!" with the
		// command name "prefix," which makes it so the actual command name is
		// empty.
		if p.Prefix == s {
			continue
		}

		if strings.HasPrefix(s, p.Prefix) {
			log.Debug().Interface("prefix", p).Msg("matched prefix")
			return p, true
		}
	}

	return Prefix{}, false
}

func (m *Message) matchPrefix(rootCmdName string) (Prefix, error) {
	prefixes, _, err := m.Prefixes()
	if err != nil {
		return Prefix{}, err
	}

	if prefix, ok := hasPrefix(prefixes, rootCmdName); ok {
		return prefix, nil
	}

	log.Debug().Str("command", rootCmdName).Msg("failed to match prefix")

	return Prefix{}, fmt.Errorf("failed to match prefix")
}

func (m *Message) CommandParse() (*Message, error) {
	log.Debug().Str("text", m.Raw).Msg("starting command parsing")

	args := m.Fields()

	if len(args) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	rootCmdName := args[0]
	prefix, err := m.matchPrefix(rootCmdName)
	if err != nil {
		return nil, err
	}
	args[0] = strings.TrimPrefix(rootCmdName, prefix.Prefix)

	cmdStatic, index, err := Commands.Match(prefix.Type, m, args)
	if err != nil {
		return nil, fmt.Errorf("couldn't match command: %v", err)
	}
	cmdName := args[:index+1]
	args = args[index+1:]

	log.Debug().
		Strs("command", cmdName).
		Strs("args", args).
		Send()

	cmdRuntime := CommandRuntime{
		Path:   cmdName,
		Args:   args,
		Prefix: prefix.Prefix,
	}

	m.Command = &Command{
		cmdStatic,
		cmdRuntime,
	}

	return m, nil
}

func (m *Message) CommandRun() (*Message, error) {
	m, err := m.CommandParse()
	if err != nil {
		return nil, err
	}

	admin, err := m.Author.BotAdmin()
	if err != nil {
		return nil, err
	}
	if m.Command.Type() == Admin && admin == false {
		return nil, fmt.Errorf("admin only command, caller not admin")
	}

	resp, urr, err := m.Command.Run(m)
	if err == UrrSilence {
		return nil, err
	}
	if err != nil {
		// passing an empty error in order to get any error-specific rendering
		// that might be supported
		if _, err := m.Write("Something went wrong...", errors.New("")); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to run command '%v': %v", m.Command.Path, err)
	}

	return m.Write(resp, urr)
}

func (m *Message) Usage() any {
	return m.Frontend.Usage(m.Command.Usage())
}
