package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// There's 2 types of errors. One type concerns the developer (something
// unexpected happened during execution) the other concerns the user (the user
// did something incorrectly). The user facing error is returned in order to
// allow special handling of error messages (for example using a different
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

	// Person returns the target's scope. If it doesn't exist it will create it
	// and add it to the database.
	Person(id string) (person int64, err error)

	// PlaceExact returns the exact scope of the specified ID.
	PlaceExact(id string) (place int64, err error)

	// PlaceLogical returns the logical scope of the specified ID.
	PlaceLogical(id string) (place int64, err error)

	// Send sends a message to the appropriate scope, resp could be nil
	// depending on the frontend.
	Send(msg any, usrErr error) (resp *Message, err error)

	// Ping works the same as Send except the user is also pinged.
	Ping(msg any, usrErr error) (resp *Message, err error)

	// Write either calls Send or Ping depending on the frontend. This is what
	// should be used in most cases.
	Write(msg any, usrErr error) (resp *Message, err error)

	// Natural will try to emulate a response as if an actual human had written
	// it. Often the bot uses markers to distinguish its responses (for example,
	// on Twitch it replies with the following format: @person -> <resp>), which
	// are not natural looking. To add to the effect, randomness may be used to
	// only sometimes mention the person.
	Natural(msg any, usrErr error) (resp *Message, err error)
}

// Author is the interface used to abstract a frontend's message author.
type Author interface {
	// ID returns the author's ID, this should be a unique, static, identifier
	// in that frontend.
	ID() (string, error)

	// Name returns the author's username.
	Name() (string, error)

	// DisplayName return's the author's display name. If only usernames exist
	// for that frontend then returns the username.
	DisplayName() (string, error)

	// Mention return's a string that mention's the author. This should ideally
	// ping them in some way.
	Mention() (string, error)

	// BotAdmin returns true if the author is a bot admin, otherwise returns
	// false.
	BotAdmin() (bool, error)

	// Admin checks if the author is considered an admin. Should return true
	// only if the author has basically every permission.
	Admin() (bool, error)

	// Mod checks if the author is considered a moderator. General rule of thumb
	// is that if the author can ban people, then they are mods.
	Mod() (bool, error)

	// Subscriber returns true if the author is considered a subscriber. General
	// rule of thumb is that if they are paying money in some way, then they are
	// a subscriber. If no such thing exists for the specific frontend, then it
	// should always return false.
	Subscriber() (bool, error)

	// Scope return's the author's scope. If it doesn't exist it will create it
	// and add it to the database.
	Scope() (author int64, err error)
}

// Here is the interface used to abstract the place where message came from, e.g.
// channel, server, etc.
//
// Two type's of scopes exist for places, the exact and the logical. The logical
// is the area where things are generally expected to work. For example: if a
// user adds a custom command in a discord server they would probably expect it
// to work in the entire server and not just in the specific channel that they
// added it in. If on the other hand someone adds a custom command in a discord
// DM, then no guild exists and thus the channel's scope would have to be used.
// On the other hand the exact scope is, as its name suggests, the scope of the
// exact place the message came from and does not account for context, so using
// the previous discord server example, it would be the channel's scope where
// the message came from instead of the server's.
type Here interface {
	// ID returns the channel's ID, this should be a unique, static, identifier
	// in that frontend.
	ID() string

	// Name return's the channel's name.
	Name() string

	// ScopeExact returns the here's exact scope. See the interface's doc
	// comment for more information on exact scopes.
	ScopeExact() (place int64, err error)

	// ScopeLogical returns the here's logical scope. See the interface's doc
	// comment for more information on logical scopes.
	ScopeLogical() (place int64, err error)
}

type Message struct {
	ID       string
	Raw      string
	Frontend Frontender
	Author   Author
	Here     Here
	Client   Messenger
	Speaker  AudioSpeaker
	Command  *Command
}

func (m *Message) Fields() []string {
	return strings.Fields(m.Raw)
}

// FieldsSpace splits text into fields that include all trailing whitespace. For
// example: "example of    text" will be split into ["example ", "of    ", "text"]
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
func (m *Message) Write(msg any, usrErr error) (*Message, error) {
	return m.Client.Write(msg, usrErr)
}

// Prefixes returns the logical here's prefixes and also whether they were taken
// from the database (if not then that means the default ones were used).
func (m *Message) Prefixes() ([]Prefix, bool, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, false, err
	}
	return PlacePrefixes(here)
}

func hasPrefix(prefixes []Prefix, s string) (Prefix, bool) {
	for _, p := range prefixes {
		// Example:
		// !prefix add !prefix
		// !prefixprefix ls // works
		// !prefix ls // doesn't work
		//
		// This is because the rootCmdName "!prefix" in the third command gets
		// matched as the prefix "!prefix" and not the prefix "!" with the
		// command name "prefix". Which makes it so the actual command name is
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
		return nil, fmt.Errorf("Empty message")
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

	resp, usrErr, err := m.Command.Run(m)
	if err == ErrSilence {
		return nil, err
	}
	if err != nil {
		// passing an empty error in order to get any error specific rendering
		// that might be supported
		m.Write("Something went wrong...", errors.New(""))
		return nil, fmt.Errorf("failed to run command '%v': %v", m.Command.Path, err)
	}

	return m.Write(resp, usrErr)
}

func (m *Message) Usage() any {
	return m.Frontend.Usage(m.Command.Usage())
}
