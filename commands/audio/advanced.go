package audio

import (
	"fmt"
	"strings"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.EventMessage) bool {
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

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (advanced) Examples() []string {
	return nil
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

func (advanced) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c advancedPlay) Permitted(m *core.EventMessage) bool {
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

func (c advancedPlay) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPlay) Examples() []string {
	return nil
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

func (c advancedPlay) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPlay) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	item, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr, discord.PlaceInBackticks(item.Title)),
	}
	return embed, urr, nil
}

func (c advancedPlay) text(m *core.EventMessage) (string, core.Urr, error) {
	item, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	title := fmt.Sprintf("'%s'", item.Title)
	return c.fmt(urr, title), urr, nil
}

func (advancedPlay) fmt(urr core.Urr, title string) string {
	switch urr {
	case nil:
		return fmt.Sprintf("Added %s in the queue.", title)
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedPlay) core(m *core.EventMessage) (Item, core.Urr, error) {
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

func (c advancedPause) Permitted(m *core.EventMessage) bool {
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

func (c advancedPause) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedPause) Examples() []string {
	return nil
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

func (c advancedPause) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedPause) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedPause) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedPause) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Paused playing."
	case UrrNotPlaying:
		return "Can't pause, nothing is playing."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedPause) core(m *core.EventMessage) (core.Urr, error) {
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

func (c advancedResume) Permitted(m *core.EventMessage) bool {
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

func (c advancedResume) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedResume) Examples() []string {
	return nil
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

func (c advancedResume) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedResume) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedResume) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedResume) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Resumed playing."
	case UrrNotPlaying:
		return "Can't resume, not playing anything."
	case UrrNotPaused:
		return "It's not paused what do you want from meeeeeeee"
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedResume) core(m *core.EventMessage) (core.Urr, error) {
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

func (c advancedSkip) Permitted(m *core.EventMessage) bool {
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

func (c advancedSkip) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSkip) Examples() []string {
	return nil
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

func (c advancedSkip) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSkip) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedSkip) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedSkip) fmt(urr error) string {
	switch urr {
	case nil:
		return "Skipped."
	case UrrNotPlaying:
		return "Can't skip, not playing anything."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedSkip) core(m *core.EventMessage) (error, error) {
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

func (c advancedLoop) Permitted(m *core.EventMessage) bool {
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

func (c advancedLoop) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedLoop) Examples() []string {
	return nil
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

func (advancedLoop) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c advancedLoopOn) Permitted(m *core.EventMessage) bool {
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

func (c advancedLoopOn) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedLoopOn) Examples() []string {
	return nil
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

func (c advancedLoopOn) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedLoopOn) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedLoopOn) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedLoopOn) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Current item is now on loop!"
	case UrrNotPlaying:
		return "Can't loop, nothing is playing."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedLoopOn) core(m *core.EventMessage) (core.Urr, error) {
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

func (c advancedLoopOff) Permitted(m *core.EventMessage) bool {
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

func (c advancedLoopOff) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedLoopOff) Examples() []string {
	return nil
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

func (c advancedLoopOff) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedLoopOff) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(urr),
	}
	return embed, urr, nil
}

func (c advancedLoopOff) text(m *core.EventMessage) (string, core.Urr, error) {
	urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(urr), urr, nil
}

func (advancedLoopOff) fmt(urr core.Urr) string {
	switch urr {
	case nil:
		return "Turned looping off."
	case UrrNotPlaying:
		return "Can't turn looping off, nothing is playing."
	case UrrNotLooping:
		return "Not looping, can't turn it off."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedLoopOff) core(m *core.EventMessage) (core.Urr, error) {
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

func (c advancedQueue) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (advancedQueue) Names() []string {
	return append([]string{"queue"}, core.AliasesList...)
}

func (advancedQueue) Description() string {
	return "View the playback queue."
}

func (advancedQueue) UsageArgs() string {
	return ""
}

func (c advancedQueue) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedQueue) Examples() []string {
	return nil
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

func (c advancedQueue) Run(m *core.EventMessage) (any, core.Urr, error) {
	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedQueue) discord(m *core.EventMessage) (*dg.MessageEmbed, core.Urr, error) {
	titles, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	if urr != nil {
		return &dg.MessageEmbed{Description: c.fmt(urr)}, urr, nil
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

func (c advancedQueue) text(m *core.EventMessage) (string, core.Urr, error) {
	titles, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if urr != nil {
		return c.fmt(urr), urr, nil
	}
	return strings.Join(titles, "  ||  "), nil, nil
}

func (advancedQueue) fmt(urr core.Urr) string {
	switch urr {
	case UrrNotPlaying:
		return "Not playing anything, the queue is empty."
	default:
		return fmt.Sprint(urr)
	}
}

func (advancedQueue) core(m *core.EventMessage) ([]string, core.Urr, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, nil, err
	}

	items, urr := Queue(here)
	if urr != nil {
		return nil, urr, nil
	}

	var titles []string
	for _, item := range items {
		titles = append(titles, item.Title)
	}

	return titles, nil, nil
}
