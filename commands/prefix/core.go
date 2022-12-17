package prefix

import (
	"errors"
	"fmt"
	"strings"

	"github.com/janitorjeff/jeff-bot/commands/custom-command"
	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	ErrExists   = errors.New("The prefix already exists.")
	ErrNotFound = errors.New("The prefix was not found.")
	ErrOneLeft  = errors.New("Only one prefix is left.")

	ErrCustomCommandExists = errors.New("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.")
)

/////////
//     //
// add //
//     //
/////////

// if the prefix changes after a custom command has been added it's
// possible that a collision may be created
//
// for example:
// !prefix reset
// !cmd add .prefix test // this works because . is not a valid prefix atm
// !prefix add .
// .prefix // both trigger
func customCommandCollision(prefix string, place int64) (string, error) {
	triggers, err := custom_command.List(place)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		if strings.HasPrefix(t, prefix) {
			return t, nil
		}
	}

	return "", nil
}

// Add adds a prefix of type t in the specified place. If the prefix, regardless
// of type, already exists then returns an ErrExists error. If a custom command
// happens to collide with the newly added prefix (for example if the custom
// command `.help` exists and the prefix `.` is added then `.help` would collide
// with the built-in command `help`) then returns an ErrCustomCommandExists
// error.
func Add(prefix string, t core.CommandType, place int64) (string, error, error) {
	prefixes, inDB, err := core.PlacePrefixes(place)
	if err != nil {
		return "", nil, err
	}

	// Only add default prefixes if they've never been added before, this
	// prevents situations were the default prefixes change and they sneakily
	// get added without the user realizing.
	if !inDB {
		for _, p := range prefixes {
			if err = dbAdd(p.Prefix, p.Type, place); err != nil {
				return "", nil, err
			}
		}
	}

	for _, p := range prefixes {
		if p.Prefix == prefix {
			return "", ErrExists, nil
		}
	}

	collision, err := customCommandCollision(prefix, place)
	if err != nil {
		return "", nil, err
	}
	if collision != "" {
		return collision, ErrCustomCommandExists, nil
	}

	return "", nil, dbAdd(prefix, t, place)
}

func cmdAdd(t core.CommandType, m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return cmdAddDiscord(t, m)
	default:
		return cmdAddText(t, m)
	}
}

func cmdAddDiscord(t core.CommandType, m *core.Message) (*dg.MessageEmbed, error, error) {
	prefix, collision, usrErr, err := cmdAddCore(t, m)
	if err != nil {
		return nil, usrErr, err
	}

	prefix = discord.PlaceInBackticks(prefix)
	collision = discord.PlaceInBackticks(collision)

	embed := &dg.MessageEmbed{
		Description: cmdAddErr(usrErr, prefix, collision),
	}

	return embed, usrErr, nil
}

func cmdAddText(t core.CommandType, m *core.Message) (string, error, error) {
	prefix, collision, usrErr, err := cmdAddCore(t, m)
	if err != nil {
		return "", usrErr, err
	}
	return cmdAddErr(usrErr, prefix, collision), usrErr, nil
}

func cmdAddErr(usrErr error, prefix, collision string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case ErrExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case ErrCustomCommandExists:
		return fmt.Sprintf(fmt.Sprint(usrErr), prefix, collision)
	default:
		return fmt.Sprint(usrErr)
	}
}

func cmdAddCore(t core.CommandType, m *core.Message) (string, string, error, error) {
	prefix := m.Command.Args[0]

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return prefix, "", nil, err
	}

	collision, usrErr, err := Add(prefix, t, here)
	return prefix, collision, usrErr, err
}

////////////
//        //
// delete //
//        //
////////////

// Delete removes the prefix of type t from the specified place. If the prefix
// doesn't exist returns an ErrNotFound error. If there is only one prefix of
// that type left then returns an ErrOneLeft error.
func Delete(prefix string, t core.CommandType, place int64) (error, error) {
	prefixes, inDB, err := core.PlacePrefixes(place)
	if err != nil {
		return nil, err
	}

	exists := false
	for _, p := range prefixes {
		if t&p.Type == 0 {
			continue
		}

		if p.Prefix == prefix {
			exists = true
		}
	}

	if !exists {
		return ErrNotFound, nil
	}
	if len(prefixes) == 1 {
		return ErrOneLeft, nil
	}

	// If the scope doesn't exist then the default prefixes are being used and
	// they are not present in the DB. So if the user tries to delete one
	// nothing will happen. So we first add them all to the DB.
	if !inDB {
		for _, p := range prefixes {
			if err = dbAdd(p.Prefix, p.Type, place); err != nil {
				return nil, err
			}
		}
	}

	return nil, dbDelete(prefix, place)
}

func cmdDelete(t core.CommandType, m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return cmdDeleteDiscord(t, m)
	default:
		return cmdDeleteText(t, m)
	}
}

func cmdDeleteDiscord(t core.CommandType, m *core.Message) (*dg.MessageEmbed, error, error) {
	prefix, usrErr, err := cmdDeleteCore(t, m)
	if err != nil {
		return nil, usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case ErrOneLeft:
		resetCommand = core.Format(AdvancedReset, m.Command.Prefix)
		resetCommand = discord.PlaceInBackticks(resetCommand)
	}

	prefix = discord.PlaceInBackticks(prefix)

	embed := &dg.MessageEmbed{
		Description: cmdDeleteErr(usrErr, m, prefix, resetCommand),
	}

	return embed, usrErr, nil
}

func cmdDeleteText(t core.CommandType, m *core.Message) (string, error, error) {
	prefix, usrErr, err := cmdDeleteCore(t, m)
	if err != nil {
		return "", usrErr, err
	}

	resetCommand := ""

	switch usrErr {
	case core.ErrMissingArgs:
		return m.Usage().(string), usrErr, nil
	case ErrOneLeft:
		resetCommand = core.Format(AdvancedReset, m.Command.Prefix)
	}

	return cmdDeleteErr(usrErr, m, prefix, resetCommand), usrErr, nil
}

func cmdDeleteErr(err error, m *core.Message, prefix, resetCommand string) string {
	switch err {
	case nil:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case ErrNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case ErrOneLeft:
		return fmt.Sprintf("Can't delete, %s is the only prefix left.\n", prefix) +
			fmt.Sprintf("If you wish to reset to the default prefixes run: %s", resetCommand)
	default:
		return "Something went wrong..."
	}
}

func cmdDeleteCore(t core.CommandType, m *core.Message) (string, error, error) {
	prefix := m.Command.Args[0]

	here, err := m.Here.ScopeLogical()
	if err != nil {
		return prefix, nil, err
	}

	usrErr, err := Delete(prefix, t, here)
	return prefix, usrErr, err
}

//////////
//      //
// list //
//      //
//////////

// List returns a list of prefixes with type t in the specified place. It is
// possible to return prefixes of multiple types like so:
//
//	List(core.Normal|core.Advanced, place)
func List(t core.CommandType, place int64) ([]core.Prefix, error) {
	prefixes, _, err := core.PlacePrefixes(place)

	ps := []core.Prefix{}
	for _, p := range prefixes {
		if t&p.Type != 0 {
			ps = append(ps, p)
		}
	}

	return ps, err
}

func cmdList(t core.CommandType, m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return cmdListDiscord(t, m)
	default:
		return cmdListText(t, m)
	}
}

func cmdListDiscord(t core.CommandType, m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	prefixes, err := cmdListCore(t, m)
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

func cmdListText(t core.CommandType, m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	prefixes, err := cmdListCore(t, m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes: %s", strings.Join(prefixes, " ")), nil, nil
}

func cmdListCore(t core.CommandType, m *core.Message) ([]string, error) {
	here, err := m.Here.ScopeLogical()
	if err != nil {
		return nil, err
	}

	prefixes, err := List(t, here)
	if err != nil {
		return nil, err
	}

	filtered := []string{}
	for _, p := range prefixes {
		if p.Type == t {
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

// Reset wipes all the prefixes from the database making it so the default ones
// will have to be used.
func Reset(place int64) error {
	return dbReset(place)
}

func cmdReset(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return cmdResetDiscord(m)
	default:
		return cmdResetText(m)
	}
}

func cmdResetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	log.Debug().Msg("running discord renderer")

	listCmd, err := cmdResetCore(m)
	if err != nil {
		return nil, nil, err
	}
	listCmd = discord.PlaceInBackticks(listCmd)

	embed := &dg.MessageEmbed{
		Description: fmt.Sprintf("Prefixes have been reset.\nTo view the list of the currently available prefixes run: %s", listCmd),
	}

	return embed, nil, nil
}

func cmdResetText(m *core.Message) (string, error, error) {
	log.Debug().Msg("running plain text renderer")

	listCmd, err := cmdResetCore(m)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("Prefixes have been reset. To view the list of the currently available prefixes run: %s", listCmd), nil, nil
}

func cmdResetCore(m *core.Message) (string, error) {
	here, err := m.Here.ScopeLogical()
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
