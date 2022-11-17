package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func checkTriggerExists(m *core.Message, trigger string) (bool, int64, error) {
	exists := false

	here, err := m.HereLogical()
	if err != nil {
		return exists, here, err
	}

	triggers, err := dbList(here)
	if err != nil {
		return exists, here, err
	}

	for _, t := range triggers {
		if t == trigger {
			exists = true
			break
		}
	}

	return exists, here, nil
}

var (
	errTriggerExists   = errors.New("trigger already exists")
	errTriggerNotFound = errors.New("trigger was not found")
	errBuiltinCommand  = errors.New("trigger collides with a built-in command")
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
		"command",
		"cmd",
	}
}

func (advanced) Description() string {
	return "Add, edit, delete or list custom commands."
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
		AdvancedEdit,
		AdvancedDelete,
		AdvancedList,
		AdvancedHistory,
	}
}

func (c advanced) Init() error {
	core.Hooks.Register(c.writeCustomCommand)
	return core.DB.Init(dbShema)
}

func (advanced) writeCustomCommand(m *core.Message) {
	fields := m.Fields()

	if len(fields) > 1 {
		return
	}

	here, err := m.HereLogical()
	if err != nil {
		return
	}

	resp, err := dbGetResponse(here, fields[0])
	if err != nil {
		return
	}

	m.Write(resp, nil)
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
	return "Add a command."
}

func (advancedAdd) UsageArgs() string {
	return "<trigger> <text>"
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

func (c advancedAdd) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func (c advancedAdd) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (advancedAdd) err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been added.", trigger)
	case errTriggerExists:
		return fmt.Sprintf("Custom command %s already exists.", trigger)
	case errBuiltinCommand:
		return fmt.Sprintf("Command %s already exists as a built-in command.", trigger)
	default:
		return "Something went wrong..."
	}
}

func (c advancedAdd) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, place, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == true {
		log.Debug().
			Int64("place", place).
			Str("trigger", trigger).
			Msg("trigger already exists in this place")

		return trigger, errTriggerExists, nil
	}

	builtin, err := c.isBuiltin(m, place, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if builtin == true {
		return trigger, errBuiltinCommand, nil
	}

	response := m.RawArgs(1)

	author, err := m.Author()
	if err != nil {
		return trigger, nil, err
	}

	err = dbAdd(place, author, trigger, response)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Str("response", response).
		Int64("author", author).
		Msg("added custom command")

	return trigger, nil, err
}

func (advancedAdd) isBuiltin(m *core.Message, place int64, trigger string) (bool, error) {
	prefixes, _, err := m.Prefixes()
	if err != nil {
		return false, err
	}

	for _, p := range prefixes {
		cmdName := []string{strings.TrimPrefix(trigger, p.Prefix)}
		_, _, err := core.Commands.Match(core.Advanced, m, cmdName)
		// if there is no error that means a command was matched and thus a
		// collision exists
		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

//////////
//      //
// edit //
//      //
//////////

var AdvancedEdit = advancedEdit{}

type advancedEdit struct{}

func (c advancedEdit) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedEdit) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedEdit) Names() []string {
	return core.Edit
}

func (advancedEdit) Description() string {
	return "Edit a command."
}

func (advancedEdit) UsageArgs() string {
	return "<trigger> <text>"
}

func (advancedEdit) Parent() core.CommandStatic {
	return Advanced
}

func (advancedEdit) Children() core.CommandsStatic {
	return nil
}

func (advancedEdit) Init() error {
	return nil
}

func (c advancedEdit) Run(m *core.Message) (any, error, error) {
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

func (c advancedEdit) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func (c advancedEdit) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (advancedEdit) err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been modified.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func (advancedEdit) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, place, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == false {
		return trigger, errTriggerNotFound, nil
	}

	response := m.RawArgs(1)

	author, err := m.Author()
	if err != nil {
		return trigger, nil, err
	}

	err = dbModify(place, author, trigger, response)

	log.Debug().
		Err(err).
		Int64("place", place).
		Int64("author", author).
		Str("trigger", trigger).
		Str("response", response).
		Msg("modified  custom command")

	return trigger, nil, err
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
	return "Delete a command."
}

func (advancedDelete) UsageArgs() string {
	return "<trigger>"
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
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: c.err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func (c advancedDelete) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (advancedDelete) err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been deleted.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func (advancedDelete) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, place, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == false {
		return trigger, errTriggerNotFound, nil
	}

	author, err := m.Author()
	if err != nil {
		return trigger, nil, err
	}

	err = dbDel(place, author, trigger)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Int64("author", author).
		Msg("deleted custom command")

	return trigger, nil, err
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
	return "List commands."
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
	triggers, err := c.Core(m)
	if err != nil {
		return nil, nil, err
	}

	var reply string

	if len(triggers) == 0 {
		reply = "There are no custom commands."
	} else {
		for i := range triggers {
			triggers[i] = "- " + discord.PlaceInBackticks(triggers[i])
		}
		reply = strings.Join(triggers, "\n")
	}

	embed := &dg.MessageEmbed{
		Description: reply,
	}

	return embed, nil, nil
}

func (c advancedList) text(m *core.Message) (string, error, error) {
	triggers, err := c.Core(m)
	if err != nil {
		return "", nil, err
	}

	if len(triggers) == 0 {
		return "There are no custom commands.", nil, nil
	}
	return strings.Join(triggers, ", "), nil, nil
}

func (c advancedList) Core(m *core.Message) ([]string, error) {
	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	triggers, err := dbList(here)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Err(err).
		Int64("place", here).
		Strs("triggers", triggers).
		Msg("got triggers")

	return triggers, nil
}

/////////////
//         //
// history //
//         //
/////////////

var AdvancedHistory = advancedHistory{}

type advancedHistory struct{}

func (c advancedHistory) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedHistory) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedHistory) Names() []string {
	return []string{
		"history",
	}
}

func (advancedHistory) Description() string {
	return "View a command's entire history of changes."
}

func (advancedHistory) UsageArgs() string {
	return "<trigger>"
}

func (advancedHistory) Parent() core.CommandStatic {
	return Advanced
}

func (advancedHistory) Children() core.CommandsStatic {
	return nil
}

func (advancedHistory) Init() error {
	return nil
}

func (c advancedHistory) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return nil, nil, nil
	}
}

func formatTime(timestamp int64) string {
	seconds := timestamp / int64(time.Second) // nanoseconds to seconds
	return fmt.Sprintf("on <t:%d:D> at <t:%d:T>", seconds, seconds)
}

func formatCreate(timestamp int64, response string) string {
	when := formatTime(timestamp)
	return fmt.Sprintf("created '%s' %s by @", response, when)
}

func formatModify(timestamp int64, response string) string {
	when := formatTime(timestamp)
	return fmt.Sprintf("modified to '%s' %s by @", response, when)
}

func formatDelete(timestamp int64) string {
	when := formatTime(timestamp)
	return fmt.Sprintf("deleted %s by @", when)
}

func (c advancedHistory) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, history, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	if len(history) == 0 {
	}

	const zeroWidthSpace = "\u200b"

	var action []string
	var response []string
	var when []string

	for i := 0; i < len(history); i++ {
		hist := history[i]

		if i == 0 {
			// creation
			action = append(action, "created")
			response = append(response, hist.response)
			when = append(when, formatTime(hist.created))
		} else if history[i-1].deleted == hist.created {
			// modification
			action = append(action, "modified")
			response = append(response, hist.response)
			when = append(when, formatTime(hist.created))
		} else {
			// deletion
			action = append(action, "deleted")
			response = append(response, history[i-1].response)
			when = append(when, formatTime(history[i-1].deleted))

			action = append(action, "created")
			response = append(response, hist.response)
			when = append(when, formatTime(hist.created))
		}

		if i == len(history)-1 && hist.deleted != 0 {
			action = append(action, "deleted")
			response = append(response, hist.response)
			when = append(when, formatTime(hist.deleted))
		}
	}

	embed := &dg.MessageEmbed{
		Title: discord.PlaceInBackticks(trigger),
		Fields: []*dg.MessageEmbedField{
			{
				Name:   "action",
				Value:  strings.Join(action, "\n"),
				Inline: true,
			},
			{
				Name:   "response",
				Value:  strings.Join(response, "\n"),
				Inline: true,
			},
			{
				Name:   "when",
				Value:  strings.Join(when, "\n"),
				Inline: true,
			},
		},
	}

	return embed, nil, nil
}

func (advancedHistory) core(m *core.Message) (string, []customCommand, error) {
	trigger := m.Command.Args[0]

	// We don't check to see if the trigger exists since this command may be
	// used to view the history of a deleted trigger

	here, err := m.HereLogical()
	if err != nil {
		return trigger, nil, err
	}

	history, err := dbHistory(here, trigger)
	if err != nil {
		return trigger, nil, err
	}

	return trigger, history, nil
}
