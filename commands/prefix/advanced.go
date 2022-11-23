package prefix

import (
	"errors"
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/command"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	errExists              = errors.New("prefix exists already")
	errNotFound            = errors.New("prefix not found")
	errOneLeft             = errors.New("only one prefix left")
	errCustomCommandExists = errors.New("if this prefix is added then there will be a collision with a custom command")
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
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedAdd) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefix, collision, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	prefix = discord.PlaceInBackticks(prefix)
	collision = discord.PlaceInBackticks(collision)

	switch usrErr {
	case core.ErrMissingArgs:
		return m.Usage().(*dg.MessageEmbed), usrErr, nil
	}

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, prefix, collision),
	}

	return embed, usrErr, nil
}

func (c advancedAdd) text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefix, collision, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	switch usrErr {
	case core.ErrMissingArgs:
		return m.Usage().(string), usrErr, nil
	}

	return c.err(usrErr, prefix, collision), usrErr, nil
}

func (c advancedAdd) err(err error, prefix, collision string) string {
	switch err {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case errExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case errCustomCommandExists:
		return fmt.Sprintf("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.", prefix, collision)
	default:
		return "Something went wrong..."
	}
}

func (c advancedAdd) core(m *core.Message) (string, string, error, error) {
	if len(m.Command.Args) < 1 {
		return "", "", core.ErrMissingArgs, nil
	}
	prefix := m.Command.Args[0]

	scope, err := m.HereLogical()
	if err != nil {
		return prefix, "", nil, err
	}
	log.Debug().Int64("scope", scope).Send()

	prefixes, scopeExists, err := m.Prefixes()
	if err != nil {
		return prefix, "", nil, err
	}

	log.Debug().
		Bool("scopeExists", scopeExists).
		Msg("checked if custom prefixes have been added before for this scope")

	// Only add default prefixes if they've never been added before, this
	// prevents situations were the default prefixes change and they sneakily
	// get added without the user realizing.
	if !scopeExists {
		log.Debug().Msg("adding default prefixes to scope prefixes")

		for _, p := range prefixes {
			if p.Type != core.Advanced {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Advanced); err != nil {
				return prefix, "", nil, err
			}
		}
	}

	exists := false
	for _, p := range prefixes {
		if p.Prefix == prefix {
			exists = true
		}
	}
	if exists {
		return prefix, "", errExists, nil
	}

	collision, err := customCommandCollision(m, prefix)
	if err != nil {
		return prefix, "", nil, err
	}
	if collision != "" {
		return prefix, collision, errCustomCommandExists, nil
	}

	err = dbAdd(prefix, scope, core.Advanced)
	return prefix, "", nil, err
}

// if the prefix changes after a custom command has been added it's
// possible that a collision maybe be created
//
// for example:
// !prefix reset
// !cmd add .prefix test // this works because . is not a valid prefix atm
// !prefix add .
// .prefix // both trigger
func customCommandCollision(m *core.Message, prefix string) (string, error) {
	triggers, err := command.AdvancedList.Core(m)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		t = strings.TrimPrefix(t, prefix)
		_, _, err := core.Commands.Match(core.Advanced, m, []string{t})
		if err == nil {
			return prefix + t, nil
		}
	}

	return "", nil
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
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefix, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case core.ErrMissingArgs:
		return m.Usage().(*dg.MessageEmbed), usrErr, nil
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
	log.Debug().Msg("running plain text renderer")

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

func (advancedDelete) core(m *core.Message) (string, error, error) {
	if len(m.Command.Args) < 1 {
		return "", core.ErrMissingArgs, nil
	}
	prefix := m.Command.Args[0]

	scope, err := m.HereLogical()
	if err != nil {
		return prefix, nil, err
	}

	log.Debug().
		Str("prefix", prefix).
		Int64("scope", scope).
		Send()

	prefixes, scopeExists, err := m.Prefixes()
	if err != nil {
		return prefix, nil, err
	}
	var exists bool
	for _, p := range prefixes {
		if p.Type != core.Advanced {
			continue
		}

		if p.Prefix == prefix {
			exists = true
		}
	}

	if !exists {
		return prefix, errNotFound, nil
	}

	if len(prefixes) == 1 {
		return prefix, errOneLeft, nil
	}

	// If the scope doesn't exist then the default prefixes are being used and
	// they are not present in the DB. So if the user tries to delete one
	// nothing will happen. So we first add them all to the DB.
	if !scopeExists {
		for _, p := range prefixes {
			if p.Type != core.Advanced {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Advanced); err != nil {
				return prefix, nil, err
			}
		}
	}

	return prefix, nil, dbDel(prefix, scope)
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

func (advancedList) core(m *core.Message) ([]string, error) {
	prefixes, _, err := m.Prefixes()

	advanced := []string{}
	for _, p := range prefixes {
		if p.Type == core.Advanced {
			advanced = append(advanced, p.Prefix)
		}
	}

	return advanced, err
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
	scope, err := m.HereLogical()
	if err != nil {
		return "", err
	}

	err = dbReset(scope)
	if err != nil {
		return "", err
	}

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
