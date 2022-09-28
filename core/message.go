package core

import (
	"fmt"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type Messenger interface {
	Parse() (*Message, error)

	// Creates a scope for the current channel. If one already exists then it
	// simply returns that.
	Scope(int) (int64, error)

	// Writes a message to the current channel
	// returned *Message could be nil depending on the platform
	//
	// TODO: Handle rate limiting gracefully, give priority to mods.
	Write(interface{}) (*Message, error)
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

// TODO: add minimum number of required args
type CommandStatic struct {
	Names       []string
	Description string
	UsageArgs   string
	Target      int64
	Run         func(*Message) (interface{}, error)
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
	ID      string
	Type    int
	IsDM    bool
	Author  *Author
	Channel *Channel
	Client  Messenger
	Command *Command
	Globals *GlobalVars
}

func (m *Message) Write(msg interface{}) (*Message, error) {
	return m.Client.Write(msg)
}

func (m *Message) Scope(scope_optional ...int) (int64, error) {
	if len(scope_optional) > 1 {
		panic(fmt.Sprintf("unexpected scope: %v", scope_optional))
	}

	scope := -1
	if len(scope_optional) == 1 {
		scope = scope_optional[0]
	}

	return m.Client.Scope(scope)
}

// Returns the current scope's prefixes and also whether or not the scope
// exists in the database (meaning that if it does exist in the DB, the
// prefixes have been modified in some way by the user)
func (m *Message) ScopePrefixes() ([]string, bool, error) {
	scope, err := m.Scope()
	if err != nil {
		return nil, false, err
	}

	prefixes, err := Globals.DB.PrefixList(scope)
	if err != nil {
		return nil, false, err
	}

	log.Debug().
		Int64("scope", scope).
		Strs("prefixes", prefixes).
		Msg("scope specific prefixes")

	if len(prefixes) != 0 {
		return prefixes, true, nil
	}

	prefixes = Globals.Prefixes()

	log.Debug().
		Strs("prefixes", prefixes).
		Msg("no scope specific prefixes, using defaults")

	if m.IsDM {
		prefixes = append(prefixes, "")

		log.Debug().
			Bool("dm", m.IsDM).
			Strs("prefixes", prefixes).
			Msg("added dm specific prefix")
	}

	return prefixes, false, nil
}

func (m *Message) CommandParse(text string) (*Message, error) {
	log.Debug().Str("text", text).Msg("starting command parsing")

	prefixes, _, err := m.ScopePrefixes()
	if err != nil {
		return nil, err
	}

	args := strings.Fields(text)

	if len(args) == 0 {
		return nil, fmt.Errorf("Empty message")
	}

	rootCmdName := args[0]
	var prefix string

	for i, p := range prefixes {
		if strings.HasPrefix(rootCmdName, p) {
			prefix = p
			log.Debug().Str("prefix", p).Msg("matched prefix")
			break
		}

		if i == len(prefixes)-1 {
			log.Debug().Msg("failed to match prefix")
			return nil, fmt.Errorf("message '%s' doesn't contain one of the expected prefixes %v", text, prefixes)
		}
	}

	args[0] = strings.TrimPrefix(rootCmdName, prefix)

	cmdStatic, index, err := Globals.Commands.MatchCommand(args)
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
		Prefix: prefix,
	}

	m.Command = &Command{
		Static:  cmdStatic,
		Runtime: cmdRuntime,
	}

	return m, nil
}

func (m *Message) CommandRun() (*Message, error) {
	cmd := m.Command.Static

	resp, err := cmd.Run(m)
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%v': %v", cmd, err)
	}

	return m.Write(resp)
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
