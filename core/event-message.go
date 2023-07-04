package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// There are 2 types of errors: one type concerns the developer (something
// unexpected happened during execution) the other concerns the user (the user
// did something incorrectly). The user facing error is returned in order to
// allow special handling of error messages (for example, using a different
// embed color in discord).

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
	Send(msg any, urr error) (resp *Message, err error)

	// Ping works the same as Send except the user is also pinged.
	Ping(msg any, urr error) (resp *Message, err error)

	// Write either calls Send or Ping depending on the frontend. This is what
	// should be used in most cases.
	Write(msg any, urr error) (resp *Message, err error)

	// Natural will try to emulate a response as if an actual human had written
	// it. Often the bot uses markers to distinguish its responses (for example,
	// on Twitch it replies with the following format: @person -> <resp>), which
	// are not natural looking. To add to the effect, randomness may be used to
	// only sometimes mention the person.
	Natural(msg any, urr error) (resp *Message, err error)
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
	if err == ErrSilence {
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
