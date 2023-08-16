package god

import (
	"fmt"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.Message) bool {
	return true
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return Advanced.Description()
}

func (c normal) UsageArgs() string {
	//return c.Children().UsageOptional()
	return "<text>"
}

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Examples() []string {
	return []string{
		"20m",
		"2h30m",
		"on",
		"off",
		"",
	}
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalShow,
		NormalOn,
		NormalOff,
		NormalMood,
		NormalDefault,
		NormalRude,
		NormalSad,
	}
}

func (normal) Init() error {
	return nil
}

func (c normal) Run(m *core.Message) (any, core.Urr, error) {
	mod, err := m.Author.Moderator()
	if err != nil {
		return nil, nil, err
	}
	if !mod {
		return AdvancedTalk.Run(m)
	}

	if len(m.Command.Args) == 0 {
		here, err := m.Here.ScopeLogical()
		if err != nil {
			return nil, nil, err
		}
		on, err := ReplyOnGet(here)
		if err != nil {
			return nil, nil, err
		}
		if on {
			// this response + You can also ask me a question directly by doing
			// !god <text> or can turn auto-replying off by doing !god off
			// Interval is set to <>, to change it use !god <interval>
			return AdvancedIntervalShow.Run(m)
		}
		return c.renderOff(m)

		// Auto-replying is off. You can talk to me directly by using
		// !god <text>. You can turn auto-replying on by doing !god on
	}
	if _, err := time.ParseDuration(m.Command.Args[0]); err == nil {
		return AdvancedIntervalSet.Run(m)
	}
	return AdvancedTalk.Run(m)
}

func (c normal) renderOff(m *core.Message) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return &dg.MessageEmbed{
			Title: "Auto-replying is off.",
			Fields: []*dg.MessageEmbedField{
				{
					Name: "Hints",
					Value: "- Talk to God directly using " + core.FormatQuote(Normal, m.Command.Prefix, m.Client) +
						"\n- Turn auto-replying on using " + core.FormatQuote(NormalOn, m.Command.Prefix, m.Client),
				},
			},
		}, nil, nil

	default:
		return c.fmtOff(m), nil, nil
	}
}

func (normal) fmtOff(m *core.Message) string {
	return fmt.Sprintf("Auto-replying is off. You can talk to God directly using %s. You can turn auto-replying on using %s", core.FormatQuote(Normal, m.Command.Prefix, m.Client), core.FormatQuote(NormalOn, m.Command.Prefix, m.Client))
}

//////////
//      //
// show //
//      //
//////////

var NormalShow = normalShow{}

type normalShow struct{}

func (c normalShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalShow) Permitted(m *core.Message) bool {
	return AdvancedReplyShow.Permitted(m)
}

func (normalShow) Names() []string {
	return AdvancedReplyShow.Names()
}

func (normalShow) Description() string {
	return AdvancedReplyShow.Description()
}

func (c normalShow) UsageArgs() string {
	return AdvancedReplyShow.UsageArgs()
}

func (c normalShow) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalShow) Examples() []string {
	return nil
}

func (normalShow) Parent() core.CommandStatic {
	return Normal
}

func (normalShow) Children() core.CommandsStatic {
	return nil
}

func (normalShow) Init() error {
	return nil
}

func (normalShow) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedReplyShow.Run(m)
}

////////
//    //
// on //
//    //
////////

var NormalOn = normalOn{}

type normalOn struct{}

func (c normalOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalOn) Permitted(m *core.Message) bool {
	return AdvancedReplyOn.Permitted(m)
}

func (normalOn) Names() []string {
	return AdvancedReplyOn.Names()
}

func (normalOn) Description() string {
	return AdvancedReplyOn.Description()
}

func (c normalOn) UsageArgs() string {
	return AdvancedReplyOn.UsageArgs()
}

func (c normalOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalOn) Examples() []string {
	return nil
}

func (normalOn) Parent() core.CommandStatic {
	return Normal
}

func (normalOn) Children() core.CommandsStatic {
	return nil
}

func (normalOn) Init() error {
	return nil
}

func (normalOn) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedReplyOn.Run(m)
}

/////////
//     //
// off //
//     //
/////////

var NormalOff = normalOff{}

type normalOff struct{}

func (c normalOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalOff) Permitted(m *core.Message) bool {
	return AdvancedReplyOff.Permitted(m)
}

func (normalOff) Names() []string {
	return AdvancedReplyOff.Names()
}

func (normalOff) Description() string {
	return AdvancedReplyOff.Description()
}

func (c normalOff) UsageArgs() string {
	return AdvancedReplyOff.UsageArgs()
}

func (c normalOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalOff) Examples() []string {
	return nil
}

func (normalOff) Parent() core.CommandStatic {
	return Normal
}

func (normalOff) Children() core.CommandsStatic {
	return nil
}

func (normalOff) Init() error {
	return nil
}

func (normalOff) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedReplyOff.Run(m)
}

//////////
//      //
// mood //
//      //
//////////

var NormalMood = normalMood{}

type normalMood struct{}

func (c normalMood) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalMood) Permitted(m *core.Message) bool {
	return AdvancedMoodShow.Permitted(m)
}

func (normalMood) Names() []string {
	return AdvancedMood.Names()
}

func (normalMood) Description() string {
	return AdvancedMoodShow.Description()
}

func (normalMood) UsageArgs() string {
	return AdvancedMoodShow.UsageArgs()
}

func (c normalMood) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalMood) Examples() []string {
	return AdvancedMoodShow.Examples()
}

func (normalMood) Parent() core.CommandStatic {
	return Normal
}

func (normalMood) Children() core.CommandsStatic {
	return nil
}

func (normalMood) Init() error {
	return nil
}

func (normalMood) Run(m *core.Message) (resp any, urr core.Urr, err error) {
	return AdvancedMoodShow.Run(m)
}

/////////////
//         //
// default //
//         //
/////////////

var NormalDefault = normalDefault{}

type normalDefault struct{}

func (c normalDefault) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalDefault) Permitted(m *core.Message) bool {
	return AdvancedMoodSetDefault.Permitted(m)
}

func (normalDefault) Names() []string {
	return AdvancedMoodSetDefault.Names()
}

func (normalDefault) Description() string {
	return AdvancedMoodSetDefault.Description()
}

func (normalDefault) UsageArgs() string {
	return AdvancedMoodSetDefault.UsageArgs()
}

func (c normalDefault) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalDefault) Examples() []string {
	return AdvancedMoodSetDefault.Examples()
}

func (normalDefault) Parent() core.CommandStatic {
	return Normal
}

func (normalDefault) Children() core.CommandsStatic {
	return nil
}

func (normalDefault) Init() error {
	return nil
}

func (normalDefault) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedMoodSetDefault.Run(m)
}

//////////
//      //
// rude //
//      //
//////////

var NormalRude = normalRude{}

type normalRude struct{}

func (c normalRude) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalRude) Permitted(m *core.Message) bool {
	return AdvancedMoodSetRude.Permitted(m)
}

func (normalRude) Names() []string {
	return AdvancedMoodSetRude.Names()
}

func (normalRude) Description() string {
	return AdvancedMoodSetRude.Description()
}

func (normalRude) UsageArgs() string {
	return AdvancedMoodSetRude.UsageArgs()
}

func (c normalRude) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalRude) Examples() []string {
	return AdvancedMoodSetRude.Examples()
}

func (normalRude) Parent() core.CommandStatic {
	return Normal
}

func (normalRude) Children() core.CommandsStatic {
	return nil
}

func (normalRude) Init() error {
	return nil
}

func (normalRude) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedMoodSetRude.Run(m)
}

/////////
//     //
// sad //
//     //
/////////

var NormalSad = normalSad{}

type normalSad struct{}

func (c normalSad) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalSad) Permitted(m *core.Message) bool {
	return AdvancedMoodSetSad.Permitted(m)
}

func (normalSad) Names() []string {
	return AdvancedMoodSetSad.Names()
}

func (normalSad) Description() string {
	return AdvancedMoodSetSad.Description()
}

func (normalSad) UsageArgs() string {
	return AdvancedMoodSetSad.UsageArgs()
}

func (c normalSad) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalSad) Examples() []string {
	return AdvancedMoodSetSad.Examples()
}

func (normalSad) Parent() core.CommandStatic {
	return Normal
}

func (normalSad) Children() core.CommandsStatic {
	return nil
}

func (normalSad) Init() error {
	return nil
}

func (normalSad) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedMoodSetSad.Run(m)
}
