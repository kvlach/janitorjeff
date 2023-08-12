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
	return Advanced.Permitted(m)
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
	}
}

func (normal) Init() error {
	return nil
}

func (c normal) Run(m *core.Message) (any, core.Urr, error) {
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
		//
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
	return c.Parent().Permitted(m)
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
	return c.Parent().Permitted(m)
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
	return c.Parent().Permitted(m)
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
