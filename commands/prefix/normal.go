package prefix

import (
	"errors"
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/commands/command"
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	errExists              = errors.New("prefix exists already")
	errNotFound            = errors.New("prefix not found")
	errOneLeft             = errors.New("only one prefix left")
	errCustomCommandExists = errors.New("if this prefix is added then there will be a collision with a custom command")
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return frontends.All
}

func (normal) Permitted(*core.Message) bool {
	return true
}

func (normal) Names() []string {
	return []string{
		"prefix",
	}
}

func (normal) Description() string {
	return "Add, delete, list or reset prefixes."
}

func (normal) UsageArgs() string {
	return "(add|del) <prefix> | list | reset"
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalAdd,
		NormalDelete,
		NormalList,
		NormalReset,
	}
}

func (c normal) Init() error {
	core.Globals.Hooks.Register(c.emergencyReset)
	return nil
}

func (normal) emergencyReset(m *core.Message) {
	if m.Raw != "!!!PleaseResetThePrefixesBackToTheDefaultsThanks!!!" {
		return
	}

	resp, usrErr, err := NormalReset.Run(m)
	if err != nil {
		return
	}

	m.Write(resp, usrErr)
}

func (normal) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////
//     //
// add //
//     //
/////////

var NormalAdd = normalAdd{}

type normalAdd struct{}

func (c normalAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalAdd) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalAdd) Names() []string {
	return core.Add
}

func (normalAdd) Description() string {
	return "Add a prefix."
}

func (normalAdd) UsageArgs() string {
	return "<prefix>"
}

func (normalAdd) Parent() core.CommandStatic {
	return Normal
}

func (normalAdd) Children() core.CommandsStatic {
	return nil
}

func (normalAdd) Init() error {
	return nil
}

func (c normalAdd) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normalAdd) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalAdd) text(m *core.Message) (string, error, error) {
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

func (c normalAdd) err(err error, prefix, collision string) string {
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

func (c normalAdd) core(m *core.Message) (string, string, error, error) {
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
			if p.Type != core.Normal {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Normal); err != nil {
				return prefix, "", nil, err
			}
		}
	}

	exists, err := dbPrefixExists(prefix, scope, core.Normal)
	if err != nil {
		return prefix, "", nil, err
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

	err = dbAdd(prefix, scope, core.Normal)
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
	triggers, err := command.NormalList.Core(m)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		t = strings.TrimPrefix(t, prefix)
		_, _, err := core.Globals.Commands.Match(core.Normal, m.Frontend, []string{t})
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

var NormalDelete = normalDelete{}

type normalDelete struct{}

func (c normalDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalDelete) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalDelete) Names() []string {
	return core.Delete
}

func (normalDelete) Description() string {
	return "Delete a prefix."
}

func (normalDelete) UsageArgs() string {
	return "<prefix>"
}

func (normalDelete) Parent() core.CommandStatic {
	return Normal
}

func (normalDelete) Children() core.CommandsStatic {
	return nil
}

func (normalDelete) Init() error {
	return nil
}

func (c normalDelete) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normalDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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
		resetCommand = core.Format(NormalReset, m.Command.Prefix)
		resetCommand = discord.PlaceInBackticks(resetCommand)
	}

	prefix = discord.PlaceInBackticks(prefix)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, m, prefix, resetCommand),
	}

	return embed, usrErr, nil
}

func (c normalDelete) text(m *core.Message) (string, error, error) {
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
		resetCommand = core.Format(NormalReset, m.Command.Prefix)
	}

	return c.err(usrErr, m, prefix, resetCommand), usrErr, nil
}

func (normalDelete) err(err error, m *core.Message, prefix, resetCommand string) string {
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

func (normalDelete) core(m *core.Message) (string, error, error) {
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
		if p.Type != core.Normal {
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
			if p.Type != core.Normal {
				continue
			}

			if err = dbAdd(p.Prefix, scope, core.Normal); err != nil {
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

var NormalList = normalList{}

type normalList struct{}

func (c normalList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalList) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalList) Names() []string {
	return core.List
}

func (normalList) Description() string {
	return "List the current prefixes."
}

func (normalList) UsageArgs() string {
	return ""
}

func (normalList) Parent() core.CommandStatic {
	return Normal
}

func (normalList) Children() core.CommandsStatic {
	return nil
}

func (normalList) Init() error {
	return nil
}

func (c normalList) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normalList) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalList) text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefixes, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes: %s", strings.Join(prefixes, " ")), nil, nil
}

func (normalList) core(m *core.Message) ([]string, error) {
	prefixes, _, err := m.Prefixes()

	normal := []string{}
	for _, p := range prefixes {
		if p.Type == core.Normal {
			normal = append(normal, p.Prefix)
		}
	}

	return normal, err
}

///////////
//       //
// reset //
//       //
///////////

var NormalReset = normalReset{}

type normalReset struct{}

func (c normalReset) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalReset) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalReset) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalReset) Names() []string {
	return []string{
		"reset",
	}
}

func (normalReset) Description() string {
	return "Reset prefixes to bot defaults."
}

func (normalReset) UsageArgs() string {
	return ""
}

func (normalReset) Parent() core.CommandStatic {
	return Normal
}

func (normalReset) Children() core.CommandsStatic {
	return nil
}

func (normalReset) Init() error {
	return nil
}

func (c normalReset) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normalReset) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalReset) text(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	listCmd, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes have been reset. To view the list of the currently available prefixes run: %s", listCmd), nil, nil
}

func (c normalReset) core(m *core.Message) (string, error) {
	scope, err := m.HereLogical()
	if err != nil {
		return "", err
	}

	err = dbReset(scope)
	if err != nil {
		return "", err
	}

	var prefix string
	for _, p := range core.Globals.Prefixes.Others {
		if p.Type == core.Normal {
			prefix = p.Prefix
			break
		}
	}

	// can't just use the prefix that was used to invoke this command because
	// it might not be valid for this scope since a reset just happened
	listCmd := core.Format(NormalList, prefix)

	return listCmd, nil
}
