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
		AdvancedSubOnly,
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

/////////////
//         //
// subonly //
//         //
/////////////

var AdvancedSubOnly = advancedSubOnly{}

type advancedSubOnly struct{}

func (c advancedSubOnly) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSubOnly) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m) && m.Author.Mod()
}

func (advancedSubOnly) Names() []string {
	return []string{
		"subonly",
	}
}

func (advancedSubOnly) Description() string {
	return "The TTS will only read subs' and mods' messages."
}

func (c advancedSubOnly) UsageArgs() string {
	return c.Children().Usage()
}

func (advancedSubOnly) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSubOnly) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedSubOnlyOn,
		AdvancedSubOnlyOff,
		AdvancedSubOnlyShow,
	}
}

func (advancedSubOnly) Init() error {
	return nil
}

func (c advancedSubOnly) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////////
//            //
// subonly on //
//            //
////////////////

var AdvancedSubOnlyOn = advancedSubOnlyOn{}

type advancedSubOnlyOn struct{}

func (c advancedSubOnlyOn) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSubOnlyOn) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSubOnlyOn) Names() []string {
	return core.AliasesOn
}

func (advancedSubOnlyOn) Description() string {
	return "Toggle sub-only mode on."
}

func (advancedSubOnlyOn) UsageArgs() string {
	return ""
}

func (advancedSubOnlyOn) Parent() core.CommandStatic {
	return AdvancedSubOnly
}

func (advancedSubOnlyOn) Children() core.CommandsStatic {
	return nil
}

func (advancedSubOnlyOn) Init() error {
	return nil
}

func (c advancedSubOnlyOn) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSubOnlyOn) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedSubOnlyOn) text(m *core.Message) (string, error, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedSubOnlyOn) fmt() string {
	return "Turned sub-only mode on."
}

func (advancedSubOnlyOn) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return SubOnlySet(here, true)
}

/////////////////
//             //
// subonly off //
//             //
/////////////////

var AdvancedSubOnlyOff = advancedSubOnlyOff{}

type advancedSubOnlyOff struct{}

func (c advancedSubOnlyOff) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSubOnlyOff) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSubOnlyOff) Names() []string {
	return core.AliasesOff
}

func (advancedSubOnlyOff) Description() string {
	return "Toggle sub-only mode off."
}

func (advancedSubOnlyOff) UsageArgs() string {
	return ""
}

func (advancedSubOnlyOff) Parent() core.CommandStatic {
	return AdvancedSubOnly
}

func (advancedSubOnlyOff) Children() core.CommandsStatic {
	return nil
}

func (advancedSubOnlyOff) Init() error {
	return nil
}

func (c advancedSubOnlyOff) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSubOnlyOff) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(),
	}
	return embed, nil, nil
}

func (c advancedSubOnlyOff) text(m *core.Message) (string, error, error) {
	err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(), nil, nil
}

func (advancedSubOnlyOff) fmt() string {
	return "Turned sub-only mode off."
}

func (advancedSubOnlyOff) core(m *core.Message) error {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return err
	}
	return SubOnlySet(here, false)
}

//////////////////
//              //
// subonly show //
//              //
//////////////////

var AdvancedSubOnlyShow = advancedSubOnlyShow{}

type advancedSubOnlyShow struct{}

func (c advancedSubOnlyShow) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSubOnlyShow) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSubOnlyShow) Names() []string {
	return core.AliasesShow
}

func (advancedSubOnlyShow) Description() string {
	return "Check if sub-only is turned on or off."
}

func (advancedSubOnlyShow) UsageArgs() string {
	return ""
}

func (advancedSubOnlyShow) Parent() core.CommandStatic {
	return AdvancedSubOnly
}

func (advancedSubOnlyShow) Children() core.CommandsStatic {
	return nil
}

func (advancedSubOnlyShow) Init() error {
	return nil
}

func (c advancedSubOnlyShow) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSubOnlyShow) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	subonly, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	embed := &dg.MessageEmbed{
		Description: c.fmt(subonly),
	}
	return embed, nil, nil
}

func (c advancedSubOnlyShow) text(m *core.Message) (string, error, error) {
	subonly, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	return c.fmt(subonly), nil, nil
}

func (c advancedSubOnlyShow) fmt(subonly bool) string {
	subonlyStr := "off"
	if subonly {
		subonlyStr = "on"
	}
	return "Sub-only mode is currently " + subonlyStr + "."
}

func (advancedSubOnlyShow) core(m *core.Message) (bool, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return false, err
	}
	return SubOnlyGet(here)
}
