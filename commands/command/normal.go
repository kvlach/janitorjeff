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

	scope, err := m.HereLogical()
	if err != nil {
		return exists, scope, err
	}

	triggers, err := dbList(scope)
	if err != nil {
		return exists, scope, err
	}

	for _, t := range triggers {
		if t == trigger {
			exists = true
			break
		}
	}

	return exists, scope, nil
}

var (
	errTriggerExists   = errors.New("trigger already exists")
	errTriggerNotFound = errors.New("trigger was not found")
	errBuiltinCommand  = errors.New("trigger collides with a built-in command")
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return frontends.All
}

func (normal) Permitted(m *core.Message) bool {
	return m.Mod()
}

func (normal) Names() []string {
	return []string{
		"command",
		"cmd",
	}
}

func (normal) Description() string {
	return "Add, edit, delete or list custom commands."
}

func (normal) UsageArgs() string {
	return "(add | edit | delete | list)"
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalAdd,
		NormalEdit,
		NormalDelete,
		NormalList,
		NormalHistory,
	}
}

func (c normal) Init() error {
	core.Globals.Hooks.Register(c.writeCustomCommand)
	return core.Globals.DB.Init(dbShema)
}

func (normal) writeCustomCommand(m *core.Message) {
	fields := m.Fields()

	if len(fields) > 1 {
		return
	}

	scope, err := m.HereLogical()
	if err != nil {
		return
	}

	resp, err := dbGetResponse(scope, fields[0])
	if err != nil {
		return
	}

	m.Write(resp, nil)
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
	return "Add a command."
}

func (normalAdd) UsageArgs() string {
	return "<trigger> <text>"
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

func (c normalAdd) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalAdd) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (normalAdd) err(usrErr error, trigger string) string {
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

func (c normalAdd) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, scope, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == true {
		log.Debug().
			Int64("scope", scope).
			Str("trigger", trigger).
			Msg("trigger already exists in this scope")

		return trigger, errTriggerExists, nil
	}

	builtin, err := c.isBuiltin(m, scope, trigger)
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

	err = dbAdd(scope, author, trigger, response)

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Str("trigger", trigger).
		Str("response", response).
		Int64("author", author).
		Msg("added custom command")

	return trigger, nil, err
}

func (normalAdd) isBuiltin(m *core.Message, scope int64, trigger string) (bool, error) {
	prefixes, _, err := m.Prefixes()
	if err != nil {
		return false, err
	}

	for _, p := range prefixes {
		cmdName := []string{strings.TrimPrefix(trigger, p.Prefix)}
		_, _, err := core.Globals.Commands.Match(core.Normal, m.Frontend, cmdName)
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

var NormalEdit = normalEdit{}

type normalEdit struct{}

func (c normalEdit) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalEdit) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalEdit) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalEdit) Names() []string {
	return core.Edit
}

func (normalEdit) Description() string {
	return "Edit a command."
}

func (normalEdit) UsageArgs() string {
	return "<trigger> <text>"
}

func (normalEdit) Parent() core.CommandStatic {
	return Normal
}

func (normalEdit) Children() core.CommandsStatic {
	return nil
}

func (normalEdit) Init() error {
	return nil
}

func (c normalEdit) Run(m *core.Message) (any, error, error) {
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

func (c normalEdit) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalEdit) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (normalEdit) err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been modified.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func (normalEdit) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, scope, err := checkTriggerExists(m, trigger)
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

	err = dbModify(scope, author, trigger, response)

	log.Debug().
		Err(err).
		Int64("scope", scope).
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
	return "Delete a command."
}

func (normalDelete) UsageArgs() string {
	return "<trigger>"
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

func (c normalDelete) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (c normalDelete) text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := c.core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return c.err(usrErr, trigger), usrErr, nil
}

func (normalDelete) err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been deleted.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func (normalDelete) core(m *core.Message) (string, error, error) {
	trigger := m.Command.Args[0]

	exists, scope, err := checkTriggerExists(m, trigger)
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

	err = dbDel(scope, author, trigger)

	log.Debug().
		Err(err).
		Int64("scope", scope).
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
	return "List commands."
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

func (c normalList) text(m *core.Message) (string, error, error) {
	triggers, err := c.Core(m)
	if err != nil {
		return "", nil, err
	}

	if len(triggers) == 0 {
		return "There are no custom commands.", nil, nil
	}
	return strings.Join(triggers, ", "), nil, nil
}

func (c normalList) Core(m *core.Message) ([]string, error) {
	scope, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	triggers, err := dbList(scope)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Err(err).
		Int64("scope", scope).
		Strs("triggers", triggers).
		Msg("got triggers")

	return triggers, nil
}

/////////////
//         //
// history //
//         //
/////////////

var NormalHistory = normalHistory{}

type normalHistory struct{}

func (c normalHistory) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalHistory) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalHistory) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalHistory) Names() []string {
	return []string{
		"history",
	}
}

func (normalHistory) Description() string {
	return "View a command's entire history of changes."
}

func (normalHistory) UsageArgs() string {
	return "<trigger>"
}

func (normalHistory) Parent() core.CommandStatic {
	return Normal
}

func (normalHistory) Children() core.CommandsStatic {
	return nil
}

func (normalHistory) Init() error {
	return nil
}

func (c normalHistory) Run(m *core.Message) (any, error, error) {
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

func (c normalHistory) discord(m *core.Message) (*dg.MessageEmbed, error, error) {
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

func (normalHistory) core(m *core.Message) (string, []customCommand, error) {
	trigger := m.Command.Args[0]

	// We don't check to see if the trigger exists since this command may be
	// used to view the history of a deleted trigger

	scope, err := m.HereLogical()
	if err != nil {
		return trigger, nil, err
	}

	history, err := dbHistory(scope, trigger)
	if err != nil {
		return trigger, nil, err
	}

	return trigger, history, nil
}
