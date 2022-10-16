package core

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	dg "github.com/bwmarrin/discordgo"
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

	// Gets the target's scope. If it doesn't exist it will create it and add
	// it to the database. The scope's type (e.g. channel, user, etc.) and an
	// ID is passed in order to specify the scope. A default scope exists that
	// expects no ID and correspons to the type -1, it returns the logical
	// scope for which actions happen in; for example in discord if a command
	// is called from a guild it will return the guild's scope instead of the
	// channel's, as opposed to a command being called from DMs. If the type
	// -2 is passed then the scope of the message's author is returned and the
	// id is also expected to be empty.
	Scope(t int, id string) (scope int64, err error)

	// Writes a message to the current channel
	// returned *Message could be nil depending on the platform
	//
	// TODO: Handle rate limiting gracefully, give priority to mods.
	Write(interface{}, error) (*Message, error)
}

type Author struct {
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
	Target      int64
	Run         func(*Message) (interface{}, error, error)
	Init        func() error

	Parent   *CommandStatic
	Children Commands
}

func (cmd *CommandStatic) Format(prefix string) string {
	path := []string{}
	for cmd.Parent != nil {
		path = append([]string{cmd.Names[0]}, path...)
		cmd = cmd.Parent
	}
	path = append([]string{cmd.Names[0]}, path...)
	return fmt.Sprintf("%s%s", prefix, strings.Join(path, " "))
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
	scopeAuthor int64
	scopePlace  int64

	ID      string
	Type    int
	Raw     string
	IsDM    bool
	Author  *Author
	Channel *Channel
	Client  Messenger
	Command *Command
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

func (m *Message) Write(msg interface{}, usrErr error) (*Message, error) {
	return m.Client.Write(msg, usrErr)
}

func (m *Message) ScopePlace() (int64, error) {
	// caches the scope to avoid unecessary database queries

	if m.scopePlace != 0 {
		return m.scopePlace, nil
	}

	place, err := m.Client.Scope(-1, "")
	if err != nil {
		return -1, err
	}
	m.scopePlace = place
	return place, nil
}

func (m *Message) ScopeAuthor() (int64, error) {
	// caches the scope to avoid unecessary database queries

	if m.scopeAuthor != 0 {
		return m.scopeAuthor, nil
	}

	author, err := m.Client.Scope(-2, "")
	if err != nil {
		return -1, err
	}
	m.scopeAuthor = author
	return author, nil
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
	if len(prefixes) != 0 {
		goto END
	}

	inDB = false
	prefixes = Globals.Prefixes.Others
	log.Debug().Msg("no scope specific prefixes, using defaults")

END:
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
	scope, err := m.ScopePlace()
	if err != nil {
		return nil, false, err
	}
	return ScopePrefixes(scope)
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

	cmdStatic, index, err := cmdList.MatchCommand(args)
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
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%v': %v", cmd, err)
	}

	return m.Write(resp, usrErr)
}

func (m *Message) Hooks() {
	for _, hook := range Globals.Hooks.Get() {
		hook(m)
	}
}

func (m *Message) Run() {
	m.Hooks()
	_, err := m.CommandRun()
	log.Debug().Err(err).Send()
}

func (m *Message) ReplyText(format string, a ...interface{}) string {
	s := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s -> %s", m.Author.Mention, s)
}

func (m *Message) replyUsageText() string {
	return m.ReplyText("Usage: %s", m.Command.Usage())
}

// FIXME: This should really not be in the core
func (m *Message) replyUsageDiscord() *dg.MessageEmbed {
	return &dg.MessageEmbed{
		Title: fmt.Sprintf("Usage: `%s`", m.Command.Usage()),
	}
}

func (m *Message) ReplyUsage() interface{} {
	switch m.Type {
	case Discord:
		return m.replyUsageDiscord()
	default:
		return m.replyUsageText()
	}
}
