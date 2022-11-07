package core

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// There's 2 types of erros. One type concerns the developer (something
// unexpected happened during execution) the other concerns the user (the user
// did something incorrectly). The user facing error is returned in order to
// allow special handling of error messages (for example using a different
// embed color in discord).

type Messenger interface {
	// Checks if the message's author is a bot admin
	Admin() bool

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

const (
	Normal = 1 << iota
	Advanced
	Admin

	All = Normal | Advanced | Admin
)

// TODO: add minimum number of required args
type CommandStatic struct {
	Names       []string
	Description string
	UsageArgs   string

	// The frontends where this command will be available at. This *must* be set
	// otherwise the command will never get matched.
	Frontends int

	Run  func(*Message) (any, error, error)
	Init func() error

	Parent   *CommandStatic
	Children Commands
}

func (cmd *CommandStatic) Format(prefix string) string {
	var args string
	if cmd.UsageArgs != "" {
		args = " " + cmd.UsageArgs
	}

	path := []string{}
	for cmd.Parent != nil {
		path = append([]string{cmd.Names[0]}, path...)
		cmd = cmd.Parent
	}
	path = append([]string{cmd.Names[0]}, path...)

	return fmt.Sprintf("%s%s%s", prefix, strings.Join(path, " "), args)
}

// type CommandRaw struct {
// 	All  string
// 	Name string
// 	Args string
// }

type CommandRuntime struct {
	// includes all the sub-commands e.g. ["prefix", "add"], so that we can
	// know which alias is being used in order to display accurate help
	// messages
	Name []string

	Args   []string
	Prefix string
	// Raw    CommandRaw
}

type Command struct {
	Type    int
	Static  *CommandStatic
	Runtime *CommandRuntime
}

func (cmd *Command) Usage() string {
	cmdName := strings.Join(cmd.Runtime.Name, " ")
	usage := fmt.Sprintf("%s%s", cmd.Runtime.Prefix, cmdName)

	if cmd.Static.UsageArgs != "" {
		usage = fmt.Sprintf("%s %s", usage, cmd.Static.UsageArgs)
	}

	return usage
}

type Message struct {
	author      int64
	hereLogical int64

	ID       string
	Frontend int
	Raw      string
	IsDM     bool
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

// Skip over first n args. Pass 0 to not skip any.
func (m *Message) RawArgs(n int) string {
	if 0 > n {
		panic("unexpected n")
	}

	fields := m.FieldsSpace()

	// Skip over the command + the given offset
	s := strings.Join(fields[len(m.Command.Runtime.Name)+n:], "")

	log.Debug().
		Int("offset", n).
		Str("raw-args", s).
		Msg("extracted raw arguments")

	return s
}

func (m *Message) Write(msg any, usrErr error) (*Message, error) {
	return m.Client.Write(msg, usrErr)
}

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

func (m *Message) HereExact() (int64, error) {
	return m.Client.PlaceExact(m.Channel.ID)
}

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

// Returns the given scope's prefixes and also whether or not they were taken
// from the database (if not then that means the default ones were used).
func ScopePrefixes(scope int64) ([]Prefix, bool, error) {
	// Initially the empty prefix was added if a message came from a DM, so
	// that normal commands could be run without using any prefix. This was
	// dropped because it added some unecessary complexity since we couldn't
	// always trivially know whether a scope was a DM or not.

	prefixes, err := Globals.DB.PrefixList(scope)
	if err != nil {
		return nil, false, err
	}

	log.Debug().
		Int64("scope", scope).
		Interface("prefixes", prefixes).
		Msg("scope specific prefixes")

	inDB := true
	if len(prefixes) == 0 {
		inDB = false
		prefixes = Globals.Prefixes.Others
		log.Debug().Msg("no scope specific prefixes, using defaults")
	}

	// The admin prefixes remain constant across scopes and can only be
	// modified through the config. This means that they are never saved in the
	// database and so we just append them to the list every time. This doesn't
	// affect the `inDB` return value.
	prefixes = append(prefixes, Globals.Prefixes.Admin...)

	// We order by the length of the prefix in order to avoid matching the
	// wrong prefix. For example, if the prefixes `!` and `!!` both exist in
	// the same scope and `!` is placed first in the list of prefixes then it
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
		Int64("scope", scope).
		Interface("prefixes", prefixes).
		Msg("got prefixes")

	return prefixes, inDB, nil
}

func (m *Message) ScopePrefixes() ([]Prefix, bool, error) {
	here, err := m.HereLogical()
	if err != nil {
		return nil, false, err
	}
	return ScopePrefixes(here)
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
	prefixes, _, err := m.ScopePrefixes()
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

	var cmdList Commands
	switch prefix.Type {
	case Normal:
		cmdList = Globals.Commands.Normal
	case Advanced:
		cmdList = Globals.Commands.Advanced
	case Admin:
		cmdList = Globals.Commands.Admin
	}

	cmdStatic, index, err := cmdList.MatchCommand(m.Frontend, args)
	if err != nil {
		return nil, fmt.Errorf("couldn't match command: %v", err)
	}
	cmdName := args[:index+1]
	args = args[index+1:]

	log.Debug().
		Strs("command", cmdName).
		Strs("args", args).
		Send()

	cmdRuntime := &CommandRuntime{
		Name:   cmdName,
		Args:   args,
		Prefix: prefix.Prefix,
	}

	m.Command = &Command{
		Type:    prefix.Type,
		Static:  cmdStatic,
		Runtime: cmdRuntime,
	}

	return m, nil
}

func (m *Message) CommandRun() (*Message, error) {
	m, err := m.CommandParse()
	if err != nil {
		return nil, err
	}

	if m.Command.Type == Admin && m.Client.Admin() == false {
		return nil, fmt.Errorf("admin only command, caller not admin")
	}

	cmd := m.Command.Static

	resp, usrErr, err := cmd.Run(m)
	if err == ErrSilence {
		return nil, err
	}
	if err != nil {
		// passing an empty error in order to get any error specific rendering
		// that might be supported
		m.Write("Something went wrong...", errors.New(""))
		return nil, fmt.Errorf("failed to run command '%v': %v", cmd, err)
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

// Monitor incoming messages until `check` is true or until timeout.
func Await(timeout time.Duration, check func(*Message) bool) *Message {
	var m *Message

	timeoutchan := make(chan bool)

	id := Globals.Hooks.Register(func(msg *Message) {
		if check(msg) {
			m = msg
			timeoutchan <- true
		}
	})

	select {
	case <-timeoutchan:
		break
	case <-time.After(timeout):
		break
	}

	Globals.Hooks.Delete(id)
	return m
}
