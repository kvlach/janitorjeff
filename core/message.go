package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// There's 2 types of erros. One type concerns the developer (something
// unexpected happened during execution) the other concerns the user (the user
// did something incorrectly). The user facing error is returned in order to
// allow special handling of error messages (for example using a different
// embed color in discord).

// The frontend abstraction layer, a frontend needs to implement this in order
// to be added.
type Messenger interface {
	// Checks if the message's author is a bot admin
	BotAdmin() bool

	Parse() (*Message, error)

	// Returns the ID of the passed string. The returned ID must be valid.
	// Generally used for verifying an ID's validity and extracting IDs from
	// mentions.
	PlaceID(s string) (id string, err error)
	PersonID(s, placeID string) (id string, err error)

	// Gets the target's scope. If it doesn't exist it will create it and add
	// it to the database.
	Person(id string) (person int64, err error)

	// There exist 2 types of place scopes that are used, the exact place and
	// the logical place. The logical is the area where things are generally
	// expected to work. For example: if a user adds a custom command in a
	// server they would probably expect it to work in the entire server and not
	// just in the specific channel that they added it in. If on the other hand
	// someone adds a custom command in a discord DM message, then no guild
	// exists and thus the channel's scope would have to be used. On the other
	// hand `PlaceExact` returns exactly the scope of the id passed and does not
	// account for context.
	PlaceExact(id string) (place int64, err error)
	PlaceLogical(id string) (place int64, err error)

	Usage(usage string) any

	// Should return true only if a user has basically every permission.
	Admin() bool

	// General rule of thumb is that if they can ban people, they are mods.
	Mod() bool

	// Sends a message to the appropriate scope, `resp` could be `nil` depending
	// on the frontend.
	Send(msg any, usrErr error) (resp *Message, err error)

	// Same as `Send` except the user is also pinged.
	Ping(msg any, usrErr error) (resp *Message, err error)

	// Either calls `Send` or `Ping` depending on the frontend. This is what
	// should be used in most cases.
	Write(msg any, usrErr error) (resp *Message, err error)
}

type User struct {
	ID          string
	Name        string
	DisplayName string
	Mention     string
}

type Channel struct {
	ID   string
	Name string
}

type Message struct {
	author      int64
	hereLogical int64

	ID       string
	Frontend int
	Raw      string
	User     *User
	Channel  *Channel
	Client   Messenger
	Command  *Command
}

func (m *Message) Fields() []string {
	return strings.Fields(m.Raw)
}

// Split text into fields that include all trailing whitespace. For example:
// "example of    text" will be split into ["example ", "of    ", "text"]
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

// Return the arguments including the whitespace between them. Skip over first
// n args. Pass 0 to not skip any.
func (m *Message) RawArgs(n int) string {
	if 0 > n {
		panic("unexpected n")
	}

	fields := m.FieldsSpace()

	// Skip over the command + the given offset
	s := strings.Join(fields[len(m.Command.Name)+n:], "")

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

// Return's the author's scope.
func (m *Message) Author() (int64, error) {
	// caches the scope to avoid unecessary database queries

	if m.author != 0 {
		return m.author, nil
	}

	author, err := m.Client.Person(m.User.ID)
	if err != nil {
		return -1, err
	}
	m.author = author
	return author, nil
}

// Return's the exact here's scope.
func (m *Message) HereExact() (int64, error) {
	return m.Client.PlaceExact(m.Channel.ID)
}

// Returns the logical here's scope.
func (m *Message) HereLogical() (int64, error) {
	// used way more often than HereExact, which is why only this one gets
	// cached

	if m.hereLogical != 0 {
		return m.hereLogical, nil
	}

	here, err := m.Client.PlaceLogical(m.Channel.ID)
	if err != nil {
		return -1, err
	}
	m.hereLogical = here
	return here, nil
}

// Returns the given place's prefixes and also whether or not they were taken
// from the database (if not then that means the default ones were used).
// Returns the logical here's prefixes.
func (m *Message) Prefixes() ([]Prefix, bool, error) {
	here, err := m.HereLogical()
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

	cmdStatic, index, err := Globals.Commands.Match(prefix.Type, m, args)
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
		Name:   cmdName,
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

	if m.Command.Type() == Admin && m.Client.BotAdmin() == false {
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
		return nil, fmt.Errorf("failed to run command '%v': %v", m.Command.Name, err)
	}

	return m.Write(resp, usrErr)
}

func (m *Message) Hooks() {
	for _, h := range Globals.Hooks.Get() {
		h.run(m)
	}
}

func (m *Message) Run() {
	m.Hooks()
	_, err := m.CommandRun()
	log.Debug().Err(err).Send()
}

func (m *Message) Usage() any {
	return m.Client.Usage(m.Command.Usage())
}

func (m *Message) Admin() bool {
	return m.Client.Admin()
}

func (m *Message) Mod() bool {
	return m.Client.Mod()
}
