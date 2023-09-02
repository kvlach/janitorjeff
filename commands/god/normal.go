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
		NormalPersonality,
		NormalPersonalities,
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
		return AdvancedTalkDialogue.Run(m)
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
			return AdvancedAutoIntervalShow.Run(m)
		}
		return c.renderOff(m)

		// Auto-replying is off. You can talk to me directly by using
		// !god <text>. You can turn auto-replying on by doing !god on
	}
	if _, err := time.ParseDuration(m.Command.Args[0]); err == nil {
		return AdvancedAutoIntervalSet.Run(m)
	}
	return AdvancedTalkDialogue.Run(m)
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
	return AdvancedAutoShow.Permitted(m)
}

func (normalShow) Names() []string {
	return AdvancedAutoShow.Names()
}

func (normalShow) Description() string {
	return AdvancedAutoShow.Description()
}

func (c normalShow) UsageArgs() string {
	return AdvancedAutoShow.UsageArgs()
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
	return AdvancedAutoShow.Run(m)
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
	return AdvancedAutoOn.Permitted(m)
}

func (normalOn) Names() []string {
	return AdvancedAutoOn.Names()
}

func (normalOn) Description() string {
	return AdvancedAutoOn.Description()
}

func (c normalOn) UsageArgs() string {
	return AdvancedAutoOn.UsageArgs()
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
	return AdvancedAutoOn.Run(m)
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
	return AdvancedAutoOff.Permitted(m)
}

func (normalOff) Names() []string {
	return AdvancedAutoOff.Names()
}

func (normalOff) Description() string {
	return AdvancedAutoOff.Description()
}

func (c normalOff) UsageArgs() string {
	return AdvancedAutoOff.UsageArgs()
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
	return AdvancedAutoOff.Run(m)
}

/////////////////
//             //
// personality //
//             //
/////////////////

var NormalPersonality = normalPersonality{}

type normalPersonality struct{}

func (c normalPersonality) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalPersonality) Permitted(m *core.Message) bool {
	return AdvancedPersonalityShow.Permitted(m)
}

func (normalPersonality) Names() []string {
	return AdvancedPersonality.Names()
}

func (normalPersonality) Description() string {
	return AdvancedPersonalityShow.Description()
}

func (normalPersonality) UsageArgs() string {
	return AdvancedPersonalityShow.UsageArgs()
}

func (c normalPersonality) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalPersonality) Examples() []string {
	return AdvancedPersonalityShow.Examples()
}

func (normalPersonality) Parent() core.CommandStatic {
	return Normal
}

func (normalPersonality) Children() core.CommandsStatic {
	return nil
}

func (normalPersonality) Init() error {
	return nil
}

func (normalPersonality) Run(m *core.Message) (any, core.Urr, error) {
	switch len(m.Command.Args) {
	case 0:
		return AdvancedPersonalityShow.Run(m)
	case 1:
		return AdvancedPersonalitySet.Run(m)
	default:
		resp, urr, err := AdvancedPersonalityAdd.Run(m)
		if err != nil {
			return nil, nil, err
		}
		if urr == UrrPersonalityExists {
			return AdvancedPersonalityEdit.Run(m)
		} else {
			return resp, urr, nil
		}
	}
}

///////////////////
//               //
// personalities //
//               //
///////////////////

var NormalPersonalities = normalPersonalities{}

type normalPersonalities struct{}

func (c normalPersonalities) Type() core.CommandType {
	return c.Parent().Type()
}

func (normalPersonalities) Permitted(m *core.Message) bool {
	return AdvancedPersonalityList.Permitted(m)
}

func (normalPersonalities) Names() []string {
	return []string{
		"personalities",
		"moods",
		"cosplays",
	}
}

func (normalPersonalities) Description() string {
	return AdvancedPersonalityList.Description()
}

func (normalPersonalities) UsageArgs() string {
	return AdvancedPersonalityList.UsageArgs()
}

func (c normalPersonalities) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalPersonalities) Examples() []string {
	return AdvancedPersonalityList.Examples()
}

func (normalPersonalities) Parent() core.CommandStatic {
	return Normal
}

func (normalPersonalities) Children() core.CommandsStatic {
	return nil
}

func (normalPersonalities) Init() error {
	return nil
}

func (normalPersonalities) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedPersonalityList.Run(m)
}
