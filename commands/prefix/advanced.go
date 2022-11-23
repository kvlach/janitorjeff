package prefix

import (
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(m *core.Message) bool {
	return m.Mod()
}

func (advanced) Names() []string {
	return []string{
		"prefix",
	}
}

func (advanced) Description() string {
	return "Add, delete, list or reset prefixes."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedAdd,
		AdvancedDelete,
		AdvancedList,
		AdvancedReset,
	}
}

func (c advanced) Init() error {
	core.Hooks.Register(c.emergencyReset)
	return nil
}

func (advanced) emergencyReset(m *core.Message) {
	if m.Raw != "!!!PleaseResetThePrefixesBackToTheDefaultsThanks!!!" {
		return
	}

	resp, usrErr, err := AdvancedReset.Run(m)
	if err != nil {
		return
	}

	m.Write(resp, usrErr)
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////
//     //
// add //
//     //
/////////

var AdvancedAdd = advancedAdd{}

type advancedAdd struct{}

func (c advancedAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedAdd) Names() []string {
	return core.Add
}

func (advancedAdd) Description() string {
	return "Add a prefix."
}

func (advancedAdd) UsageArgs() string {
	return "<prefix>"
}

func (advancedAdd) Parent() core.CommandStatic {
	return Advanced
}

func (advancedAdd) Children() core.CommandsStatic {
	return nil
}

func (advancedAdd) Init() error {
	return nil
}

func (c advancedAdd) Run(m *core.Message) (any, error, error) {
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

func (c advancedAdd) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	prefix, collision, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	prefix = discord.PlaceInBackticks(prefix)
	collision = discord.PlaceInBackticks(collision)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, prefix, collision),
	}

	return embed, usrErr, nil
}

func (c advancedAdd) text(m *core.Message) (string, error, error) {
	prefix, collision, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}
	return c.err(usrErr, prefix, collision), usrErr, nil
}

func (c advancedAdd) err(usrErr error, prefix, collision string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case errExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case errCustomCommandExists:
		return fmt.Sprintf(fmt.Sprint(usrErr), prefix, collision)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (c advancedAdd) core(m *core.Message) (string, string, error, error) {
	prefix := m.Command.Args[0]

	here, err := m.HereLogical()
	if err != nil {
		return prefix, "", nil, err
	}

	collision, usrErr, err := Add(prefix, c.Type(), here)
	return prefix, collision, usrErr, err
}

////////////
//        //
// delete //
//        //
////////////

var AdvancedDelete = advancedDelete{}

type advancedDelete struct{}

func (c advancedDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedDelete) Names() []string {
	return core.Delete
}

func (advancedDelete) Description() string {
	return "Delete a prefix."
}

func (advancedDelete) UsageArgs() string {
	return "<prefix>"
}

func (advancedDelete) Parent() core.CommandStatic {
	return Advanced
}

func (advancedDelete) Children() core.CommandsStatic {
	return nil
}

func (advancedDelete) Init() error {
	return nil
}

func (c advancedDelete) Run(m *core.Message) (any, error, error) {
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

func (c advancedDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	prefix, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case errOneLeft:
		resetCommand = core.Format(AdvancedReset, m.Command.Prefix)
		resetCommand = discord.PlaceInBackticks(resetCommand)
	}

	prefix = discord.PlaceInBackticks(prefix)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, m, prefix, resetCommand),
	}

	return embed, usrErr, nil
}

func (c advancedDelete) text(m *core.Message) (string, error, error) {
	prefix, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case core.ErrMissingArgs:
		return m.Usage().(string), usrErr, nil
	case errOneLeft:
		resetCommand = core.Format(AdvancedReset, m.Command.Prefix)
	}

	return c.err(usrErr, m, prefix, resetCommand), usrErr, nil
}

func (advancedDelete) err(err error, m *core.Message, prefix, resetCommand string) string {
	switch err {
	case nil:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case errNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case errOneLeft:
		return fmt.Sprintf("Can't delete, %s is the only prefix left.\n", prefix) +
			fmt.Sprintf("If you wish to reset to the default prefixes run: %s", resetCommand)
	default:
		return "Something went wrong..."
	}
}

func (c advancedDelete) core(m *core.Message) (string, error, error) {
	prefix := m.Command.Args[0]

	here, err := m.HereLogical()
	if err != nil {
		return prefix, nil, err
	}

	usrErr, err := Delete(prefix, c.Type(), here)
	return prefix, usrErr, err
}

//////////
//      //
// list //
//      //
//////////

var AdvancedList = advancedList{}

type advancedList struct{}

func (c advancedList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedList) Names() []string {
	return core.List
}

func (advancedList) Description() string {
	return "List the current prefixes."
}

func (advancedList) UsageArgs() string {
	return ""
}

func (advancedList) Parent() core.CommandStatic {
	return Advanced
}

func (advancedList) Children() core.CommandsStatic {
	return nil
}

func (advancedList) Init() error {
	return nil
}

func (c advancedList) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedList) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefixes, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	for i, p := range prefixes {
		// TODO: Add a warning message saying that the prefix can not be
		// displayed correctly on android devices
		prefixes[i] = discord.PlaceInBackticks(p + "command")
	}

	embed := &dg.MessageEmbed{
		Title:       "Prefixes",
		Description: strings.Join(prefixes, "\n"),
	}

	return embed, nil, nil
}

func (c advancedList) text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefixes, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes: %s", strings.Join(prefixes, " ")), nil, nil
}

func (c advancedList) core(m *core.Message) ([]string, error) {
	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	prefixes, err := List(c.Type(), here)
	if err != nil {
		return nil, err
	}

	filtered := []string{}
	for _, p := range prefixes {
		if p.Type == c.Type() {
			filtered = append(filtered, p.Prefix)
		}
	}

	return filtered, nil
}

///////////
//       //
// reset //
//       //
///////////

var AdvancedReset = advancedReset{}

type advancedReset struct{}

func (c advancedReset) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedReset) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedReset) Names() []string {
	return []string{
		"reset",
	}
}

func (advancedReset) Description() string {
	return "Reset prefixes to bot defaults."
}

func (advancedReset) UsageArgs() string {
	return ""
}

func (advancedReset) Parent() core.CommandStatic {
	return Advanced
}

func (advancedReset) Children() core.CommandsStatic {
	return nil
}

func (advancedReset) Init() error {
	return nil
}

func (c advancedReset) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedReset) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	listCmd, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	listCmd = discord.PlaceInBackticks(listCmd)

	embed := &dg.MessageEmbed{
		Description: fmt.Sprintf("Prefixes have been reset.\nTo view the list of the currently available prefixes run: %s", listCmd),
	}

	return embed, nil, nil
}

func (c advancedReset) text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	listCmd, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes have been reset. To view the list of the currently available prefixes run: %s", listCmd), nil, nil
}

func (c advancedReset) core(m *core.Message) (string, error) {
	here, err := m.HereLogical()
	if err != nil {
		return "", err
	}

	err = Reset(here)

	var prefix string
	for _, p := range core.Prefixes.Others() {
		if p.Type == core.Advanced {
			prefix = p.Prefix
			break
		}
	}

	// can't just use the prefix that was used to invoke this command because
	// it might not be valid for this scope since a reset just happened
	listCmd := core.Format(AdvancedList, prefix)

	return listCmd, nil
}
