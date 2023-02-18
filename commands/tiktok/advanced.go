package tiktok

import (
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/nick"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"tiktok",
	}
}

func (advanced) Description() string {
	return "TikTok TTS."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedStart,
		AdvancedStop,
		AdvancedUser,
	}
}

func (advanced) Init() error {
	return core.DB.Init(dbSchema)
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

///////////
//       //
// start //
//       //
///////////

var AdvancedStart = advancedStart{}

type advancedStart struct{}

func (c advancedStart) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedStart) Permitted(m *core.Message) bool {
	return m.Speaker.Enabled()
}

func (c advancedStart) Names() []string {
	return []string{
		"start",
	}
}

func (advancedStart) Description() string {
	return "Start the TTS."
}

func (advancedStart) UsageArgs() string {
	return "<twitch channel>"
}

func (advancedStart) Parent() core.CommandStatic {
	return Advanced
}

func (advancedStart) Children() core.CommandsStatic {
	return nil
}

func (advancedStart) Init() error {
	return nil
}

func (c advancedStart) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedStart) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	c.core(m)
	embed := &dg.MessageEmbed{
		Description: "Monitoring channel.",
	}
	return embed, nil, nil
}

func (c advancedStart) text(m *core.Message) (string, error, error) {
	c.core(m)
	return "Monitoring channel.", nil, nil
}

func (advancedStart) core(m *core.Message) {
	twitchUsername := strings.ToLower(m.Command.Args[0])
	Start(m.Speaker, twitchUsername)
}

//////////
//      //
// stop //
//      //
//////////

var AdvancedStop = advancedStop{}

type advancedStop struct{}

func (c advancedStop) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedStop) Permitted(m *core.Message) bool {
	return AdvancedStart.Permitted(m)
}

func (advancedStop) Names() []string {
	return []string{
		"stop",
	}
}

func (advancedStop) Description() string {
	return "Stop the TTS."
}

func (advancedStop) UsageArgs() string {
	return "<twitch channel>"
}

func (advancedStop) Parent() core.CommandStatic {
	return Advanced
}

func (advancedStop) Children() core.CommandsStatic {
	return nil
}

func (advancedStop) Init() error {
	return nil
}

func (c advancedStop) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedStop) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr := c.core(m)
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedStop) text(m *core.Message) (string, error, error) {
	usrErr := c.core(m)
	return c.err(usrErr), usrErr, nil
}

func (c advancedStop) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Stopped monitoring."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedStop) core(m *core.Message) error {
	twitchUsername := strings.ToLower(m.Command.Args[0])
	return Stop(twitchUsername)
}

//////////
//      //
// user //
//      //
//////////

var AdvancedUser = advancedUser{}

type advancedUser struct{}

func (c advancedUser) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedUser) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m) && m.Author.Mod()
}

func (advancedUser) Names() []string {
	return []string{
		"user",
	}
}

func (advancedUser) Description() string {
	return "Set a user's voice."
}

func (advancedUser) UsageArgs() string {
	return "<user> <voice>"
}

func (advancedUser) Parent() core.CommandStatic {
	return Advanced
}

func (advancedUser) Children() core.CommandsStatic {
	return nil
}

func (advancedUser) Init() error {
	return nil
}

func (c advancedUser) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 2 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedUser) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	voice, err := c.core(m)

	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: "Added voice " + voice,
	}

	return embed, nil, nil
}

func (c advancedUser) text(m *core.Message) (string, error, error) {
	voice, err := c.core(m)

	if err != nil {
		return "", nil, err
	}

	return "Added voice " + voice, nil, nil
}

func (advancedUser) core(m *core.Message) (string, error) {
	user := m.Command.Args[0]
	voice := m.Command.Args[1]

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return "", err
	}

	person, err := nick.ParsePersonHere(m, user)
	if err != nil {
		return "", err
	}

	return voice, UserVoiceSet(person, here, voice)
}
