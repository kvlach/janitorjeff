package audio

import (
	"errors"
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/youtube"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/janitorjeff/gosafe"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return m.Frontend == frontends.Discord
}

func (advanced) Names() []string {
	return []string{
		"audio",
	}
}

func (advanced) Description() string {
	return "Play music yo."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedPlay,
		AdvancedPause,
		AdvancedResume,
		AdvancedSkip,
		AdvancedQueue,
	}
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

//////////
//      //
// play //
//      //
//////////

var AdvancedPlay = advancedPlay{}

type advancedPlay struct{}

func (c advancedPlay) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPlay) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPlay) Names() []string {
	return []string{
		"play",
		"p",
	}
}

func (advancedPlay) Description() string {
	return "Add a video to the queue."
}

func (advancedPlay) UsageArgs() string {
	return "<url> | <query...>"
}

func (advancedPlay) Parent() core.CommandStatic {
	return Advanced
}

func (advancedPlay) Children() core.CommandsStatic {
	return nil
}

func (advancedPlay) Init() error {
	return nil
}

func (advancedPlay) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	url := m.Command.Args[0]

	var item Item

	if IsValidURL(url) {
		info, err := GetInfo(url)
		if err != nil {
			panic("site not supported or something else went wrong")
		}
		item = info
	} else {
		vid, usrErr, err := youtube.SearchVideo(m.RawArgs(0))
		if err != nil || usrErr != nil {
			panic(err)
		}
		item = Item{
			URL:   vid.URL(),
			Title: vid.Title,
		}
	}

	d := m.Client.(*discord.MessageCreate)
	guildID := d.Message.GuildID

	here, err := m.HereLogical()
	if err != nil {
		panic(err)
	}

	if p, ok := playing.Get(here); ok {
		p.Queue.Append(item)
		embed := &dg.MessageEmbed{
			Description: "Added to queue: " + item.URL,
		}
		return embed, nil, nil
	}

	v, err := discord.JoinUserVoiceChannel(discord.Session, guildID, m.User.ID)
	if err != nil {
		panic(err)
	}

	q := &gosafe.Slice[Item]{}
	q.Append(item)

	p := &Playing{
		State: &core.State{},
		Queue: q,
	}
	go stream(v, p, here)
	playing.Set(here, p)

	embed := &dg.MessageEmbed{
		Description: "Playing " + item.URL,
	}

	return embed, nil, nil
}

///////////
//       //
// pause //
//       //
///////////

var AdvancedPause = advancedPause{}

type advancedPause struct{}

func (c advancedPause) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedPause) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedPause) Names() []string {
	return []string{
		"pause",
	}
}

func (advancedPause) Description() string {
	return "Pause what is playing."
}

func (advancedPause) UsageArgs() string {
	return ""
}

func (advancedPause) Parent() core.CommandStatic {
	return Advanced
}

func (advancedPause) Children() core.CommandsStatic {
	return nil
}

func (advancedPause) Init() error {
	return nil
}

func (advancedPause) Run(m *core.Message) (any, error, error) {
	here, err := m.HereLogical()
	if err != nil {
		panic(err)
	}

	p, ok := playing.Get(here)

	if !ok {
		embed := &dg.MessageEmbed{
			Description: "Not playing anything, can't pause.",
		}
		return embed, fmt.Errorf("Not playing anything."), nil
	}

	if p.State.Get() == core.Play {
		p.State.Set(core.Pause)
		embed := &dg.MessageEmbed{
			Description: "Paused playing.",
		}
		return embed, nil, nil
	} else {
		embed := &dg.MessageEmbed{
			// Description: "Already paused.",
			Description: "it's not playing why are you trying to pause fool",
		}
		return embed, fmt.Errorf("not paused"), nil
	}
}

////////////
//        //
// resume //
//        //
////////////

var AdvancedResume = advancedResume{}

type advancedResume struct{}

func (c advancedResume) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedResume) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedResume) Names() []string {
	return []string{
		"resume",
		"unpause",
	}
}

func (advancedResume) Description() string {
	return "Resume playing."
}

func (advancedResume) UsageArgs() string {
	return ""
}

func (advancedResume) Parent() core.CommandStatic {
	return Advanced
}

func (advancedResume) Children() core.CommandsStatic {
	return nil
}

func (advancedResume) Init() error {
	return nil
}

func (advancedResume) Run(m *core.Message) (any, error, error) {
	here, err := m.HereLogical()
	if err != nil {
		panic(err)
	}

	p, ok := playing.Get(here)

	if !ok {
		embed := &dg.MessageEmbed{
			Description: "Not playing anything, can't resume.",
		}
		return embed, fmt.Errorf("Not playing anything."), nil

	}

	if p.State.Get() == core.Pause {
		p.State.Set(core.Play)
		embed := &dg.MessageEmbed{
			Description: "Resumed playing.",
		}
		return embed, nil, nil
	} else {
		embed := &dg.MessageEmbed{
			// Description: "It's not paused, what's the point of resuming!",
			Description: "it's not paused what do you want from meeeeeeee",
		}
		return embed, errors.New("Not paused"), nil
	}
}

//////////
//      //
// skip //
//      //
//////////

var AdvancedSkip = advancedSkip{}

type advancedSkip struct{}

func (c advancedSkip) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSkip) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSkip) Names() []string {
	return []string{
		"skip",
	}
}

func (advancedSkip) Description() string {
	return "Skip the current song."
}

func (advancedSkip) UsageArgs() string {
	return ""
}

func (advancedSkip) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSkip) Children() core.CommandsStatic {
	return nil
}

func (advancedSkip) Init() error {
	return nil
}

func (c advancedSkip) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		panic("this should never trigger")
	}
}

func (c advancedSkip) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (advancedSkip) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Skipped."
	case ErrNotPlaying:
		return "Can't skip, not playing anything."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedSkip) core(m *core.Message) (error, error) {
	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}
	return Skip(here), nil
}

///////////
//       //
// queue //
//       //
///////////

var AdvancedQueue = advancedQueue{}

type advancedQueue struct{}

func (c advancedQueue) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedQueue) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedQueue) Names() []string {
	return []string{
		"queue",
	}
}

func (advancedQueue) Description() string {
	return "View the song queue."
}

func (advancedQueue) UsageArgs() string {
	return ""
}

func (advancedQueue) Parent() core.CommandStatic {
	return Advanced
}

func (advancedQueue) Children() core.CommandsStatic {
	return nil
}

func (advancedQueue) Init() error {
	return nil
}

func (c advancedQueue) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		panic("this should never trigger")
	}
}

func (c advancedQueue) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	items, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if usrErr != nil {
		return &dg.MessageEmbed{Description: c.err(usrErr)}, usrErr, nil
	}

	var titles []string
	for _, item := range items {
		titles = append(titles, item.Title)
	}

	embed := &dg.MessageEmbed{
		Fields: []*dg.MessageEmbedField{
			{
				Name:  "Titles",
				Value: strings.Join(titles, "\n"),
			},
		},
	}

	return embed, nil, nil
}

func (advancedQueue) err(usrErr error) string {
	switch usrErr {
	case ErrNotPlaying:
		return "Not playing anything, the queue is empty."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedQueue) core(m *core.Message) ([]Item, error, error) {
	here, err := m.HereLogical()
	if err != nil {
		return nil, nil, err
	}

	items, usrErr := Queue(here)
	if usrErr != nil {
		return nil, usrErr, nil
	}

	return items, nil, nil
}
