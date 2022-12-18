package audio

import (
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return m.Speaker.Enabled()
}

func (advanced) Names() []string {
	return []string{
		"audio",
	}
}

func (advanced) Description() string {
	return "Audio related commands."
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
		AdvancedLoop,
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

func (c advancedPlay) Run(m *core.Message) (any, error, error) {
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

func (c advancedPlay) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	item, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, discord.PlaceInBackticks(item.Title)),
	}
	return embed, usrErr, nil
}

func (c advancedPlay) text(m *core.Message) (string, error, error) {
	item, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	title := fmt.Sprintf("'%s'", item.Title)
	return c.err(usrErr, title), usrErr, nil
}

func (advancedPlay) err(usrErr error, title string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added %s in the queue.", title)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedPlay) core(m *core.Message) (Item, error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return Item{}, nil, err
	}
	return Play(m.Command.Args, m.Speaker, here)
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

func (c advancedPause) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPause) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedPause) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedPause) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Paused playing."
	case ErrNotPlaying:
		return "Can't pause, nothing is playing."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedPause) core(m *core.Message) (error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return Pause(here), nil
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

func (c advancedResume) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedResume) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedResume) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedResume) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Resumed playing."
	case ErrNotPlaying:
		return "Can't resume, not playing anything."
	case ErrNotPaused:
		return "It's not paused what do you want from meeeeeeee"
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedResume) core(m *core.Message) (error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return Resume(here), nil
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
	return "Skip the currently playing item."
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
		return c.text(m)
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

func (c advancedSkip) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
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
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return Skip(here), nil
}

//////////
//      //
// loop //
//      //
//////////

var AdvancedLoop = advancedLoop{}

type advancedLoop struct{}

func (c advancedLoop) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedLoop) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedLoop) Names() []string {
	return []string{
		"loop",
	}
}

func (advancedLoop) Description() string {
	return "Turn looping on or off."
}

func (c advancedLoop) UsageArgs() string {
	return c.Children().Usage()
}

func (advancedLoop) Parent() core.CommandStatic {
	return Advanced
}

func (advancedLoop) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedLoopOn,
		AdvancedLoopOff,
	}
}

func (advancedLoop) Init() error {
	return nil
}

func (advancedLoop) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////////
//         //
// loop on //
//         //
/////////////

var AdvancedLoopOn = advancedLoopOn{}

type advancedLoopOn struct{}

func (c advancedLoopOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedLoopOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedLoopOn) Names() []string {
	return []string{
		"on",
	}
}

func (advancedLoopOn) Description() string {
	return "Will play the current item on loop, indefinitely!"
}

func (advancedLoopOn) UsageArgs() string {
	return ""
}

func (advancedLoopOn) Parent() core.CommandStatic {
	return AdvancedLoop
}

func (advancedLoopOn) Children() core.CommandsStatic {
	return nil
}

func (advancedLoopOn) Init() error {
	return nil
}

func (c advancedLoopOn) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedLoopOn) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedLoopOn) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedLoopOn) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Current item is now on loop!"
	case ErrNotPlaying:
		return "Can't loop, nothing is playing."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedLoopOn) core(m *core.Message) (error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return LoopOn(here), nil
}

//////////////
//          //
// loop off //
//          //
//////////////

var AdvancedLoopOff = advancedLoopOff{}

type advancedLoopOff struct{}

func (c advancedLoopOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedLoopOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedLoopOff) Names() []string {
	return []string{
		"off",
	}
}

func (advancedLoopOff) Description() string {
	return "Turn looping off."
}

func (advancedLoopOff) UsageArgs() string {
	return ""
}

func (advancedLoopOff) Parent() core.CommandStatic {
	return AdvancedLoop
}

func (advancedLoopOff) Children() core.CommandsStatic {
	return nil
}

func (advancedLoopOff) Init() error {
	return nil
}

func (c advancedLoopOff) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedLoopOff) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.err(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedLoopOff) text(m *core.Message) (string, error, error) {
	usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.err(usrErr), usrErr, nil
}

func (advancedLoopOff) err(usrErr error) string {
	switch usrErr {
	case nil:
		return "Turned looping off."
	case ErrNotPlaying:
		return "Can't turn looping off, nothing is playing."
	case ErrNotLooping:
		return "Not looping, can't turn it off."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedLoopOff) core(m *core.Message) (error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}
	return LoopOff(here), nil
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
	return "View the playback queue."
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
		return c.text(m)
	}
}

func (c advancedQueue) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	titles, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if usrErr != nil {
		return &dg.MessageEmbed{Description: c.err(usrErr)}, usrErr, nil
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

func (c advancedQueue) text(m *core.Message) (string, error, error) {
	titles, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr != nil {
		return c.err(usrErr), usrErr, nil
	}
	return strings.Join(titles, "  ||  "), nil, nil
}

func (advancedQueue) err(usrErr error) string {
	switch usrErr {
	case ErrNotPlaying:
		return "Not playing anything, the queue is empty."
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedQueue) core(m *core.Message) ([]string, error, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, nil, err
	}

	items, usrErr := Queue(here)
	if usrErr != nil {
		return nil, usrErr, nil
	}

	var titles []string
	for _, item := range items {
		titles = append(titles, item.Title)
	}

	return titles, nil, nil
}
