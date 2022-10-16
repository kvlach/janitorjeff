package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/platforms/discord"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	errTriggerExists   = errors.New("trigger already exists")
	errTriggerNotFound = errors.New("trigger was not found")
	errBuiltinCommand  = errors.New("trigger collides with a built-in command")
	errMissingArgs     = errors.New("not enough arguments provided")
)

func run(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), errMissingArgs, nil
}

func runAdd(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.ReplyUsage(), errMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runAdd_Discord(m)
	default:
		return runAdd_Text(m)
	}
}

func runAdd_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, usrErr, err := runAdd_Core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: runAdd_Err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func runAdd_Text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := runAdd_Core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return runAdd_Err(usrErr, trigger), usrErr, nil
}

func runAdd_Err(usrErr error, trigger string) string {
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

func runAdd_Core(m *core.Message) (string, error, error) {
	trigger := m.Command.Runtime.Args[0]

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

	builtin, err := isBuiltin(m, scope, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if builtin == true {
		return trigger, errBuiltinCommand, nil
	}

	response := m.RawArgs(1)

	author, err := m.ScopeAuthor()
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

func isBuiltin(m *core.Message, scope int64, trigger string) (bool, error) {
	prefixes, _, err := m.ScopePrefixes()
	if err != nil {
		return false, err
	}

	for _, p := range prefixes {
		cmdName := []string{strings.TrimPrefix(trigger, p.Prefix)}
		_, _, err := core.Globals.Commands.Normal.MatchCommand(cmdName)
		// if there is no error that means a command was matched and thus a
		// collision exists
		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

func runModify(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 2 {
		return m.ReplyUsage(), errMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runModify_Discord(m)
	default:
		return runModify_Text(m)
	}
}

func runModify_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, usrErr, err := runModify_Core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: runModify_Err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func runModify_Text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := runModify_Core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return runModify_Err(usrErr, trigger), usrErr, nil
}

func runModify_Err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been modified.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func runModify_Core(m *core.Message) (string, error, error) {
	trigger := m.Command.Runtime.Args[0]

	exists, scope, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == false {
		return trigger, errTriggerNotFound, nil
	}

	response := m.RawArgs(1)

	author, err := m.ScopeAuthor()
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

func runDel(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), errMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runDel_Discord(m)
	default:
		return runDel_Text(m)
	}
}

func runDel_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, usrErr, err := runDel_Core(m)
	if err != nil {
		return nil, usrErr, err
	}

	trigger = discord.PlaceInBackticks(trigger)

	embed := &dg.MessageEmbed{
		Description: runDel_Err(usrErr, trigger),
	}

	return embed, usrErr, nil
}

func runDel_Text(m *core.Message) (string, error, error) {
	trigger, usrErr, err := runDel_Core(m)
	if err != nil {
		return "", usrErr, err
	}

	trigger = fmt.Sprintf("'%s'", trigger)

	return runDel_Err(usrErr, trigger), usrErr, nil
}

func runDel_Err(usrErr error, trigger string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Custom command %s has been deleted.", trigger)
	case errTriggerNotFound:
		return fmt.Sprintf("Custom command %s doesn't exist.", trigger)
	default:
		return "Something went wrong..."
	}
}

func runDel_Core(m *core.Message) (string, error, error) {
	trigger := m.Command.Runtime.Args[0]

	exists, scope, err := checkTriggerExists(m, trigger)
	if err != nil {
		return trigger, nil, err
	}
	if exists == false {
		return trigger, errTriggerNotFound, nil
	}

	author, err := m.ScopeAuthor()
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

func checkTriggerExists(m *core.Message, trigger string) (bool, int64, error) {
	exists := false

	scope, err := m.ScopePlace()
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

func runList(m *core.Message) (any, error, error) {
	switch m.Type {
	case core.Discord:
		return runList_Discord(m)
	default:
		return runList_Text(m)
	}
}

func runList_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	triggers, err := RunList_Core(m)
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

func runList_Text(m *core.Message) (string, error, error) {
	triggers, err := RunList_Core(m)
	if err != nil {
		return "", nil, err
	}

	if len(triggers) == 0 {
		return "There are no custom commands.", nil, nil
	}
	return strings.Join(triggers, ", "), nil, nil
}

func RunList_Core(m *core.Message) ([]string, error) {
	scope, err := m.ScopePlace()
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

func runHistory(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), errMissingArgs, nil
	}

	switch m.Type {
	case core.Discord:
		return runHistory_Discord(m)
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

func runHistory_Discord(m *core.Message) (*dg.MessageEmbed, error, error) {
	trigger, history, err := runHistory_Core(m)
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

func runHistory_Core(m *core.Message) (string, []customCommand, error) {
	trigger := m.Command.Runtime.Args[0]

	// We don't check to see if the trigger exists since this command may be
	// used to view the history of a deleted trigger

	scope, err := m.ScopePlace()
	if err != nil {
		return trigger, nil, err
	}

	history, err := dbHistory(scope, trigger)
	if err != nil {
		return trigger, nil, err
	}

	return trigger, history, nil
}
