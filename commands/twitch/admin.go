package twitch

import (
	"errors"
	"strings"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"
	"github.com/rs/zerolog/log"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(m *core.Message) bool {
	return m.Frontend.Type() == twitch.Type
}

func (admin) Names() []string {
	return []string{
		"twitch",
		"ttv",
	}
}

func (admin) Description() string {
	return "Twitch related admin operations."
}

func (c admin) UsageArgs() string {
	return c.Children().Usage()
}

func (admin) Category() core.CommandCategory {
	return core.CommandCategoryOther
}

func (admin) Examples() []string {
	return nil
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminEventSub,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return m.Usage(), nil, nil
}

//////////////
//          //
// eventsub //
//          //
//////////////

var AdminEventSub = adminEventSub{}

type adminEventSub struct{}

func (c adminEventSub) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminEventSub) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminEventSub) Names() []string {
	return []string{
		"eventsub",
	}
}

func (adminEventSub) Description() string {
	return "Control EventSub."
}

func (c adminEventSub) UsageArgs() string {
	return c.Children().Usage()
}

func (c adminEventSub) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminEventSub) Examples() []string {
	return nil
}

func (adminEventSub) Parent() core.CommandStatic {
	return Admin
}

func (adminEventSub) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminEventSubList,
		AdminEventSubDelete,
	}
}

func (adminEventSub) Init() error {
	return nil
}

func (adminEventSub) Run(m *core.Message) (any, error, error) {
	return m.Usage(), nil, nil
}

///////////////////
//               //
// eventsub list //
//               //
///////////////////

var AdminEventSubList = adminEventSubList{}

type adminEventSubList struct{}

func (c adminEventSubList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminEventSubList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminEventSubList) Names() []string {
	return core.AliasesList
}

func (adminEventSubList) Description() string {
	return "List all subscriptions."
}

func (adminEventSubList) UsageArgs() string {
	return ""
}

func (c adminEventSubList) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminEventSubList) Examples() []string {
	return nil
}

func (adminEventSubList) Parent() core.CommandStatic {
	return AdminEventSub
}

func (adminEventSubList) Children() core.CommandsStatic {
	return nil
}

func (adminEventSubList) Init() error {
	return nil
}

func (adminEventSubList) Run(*core.Message) (any, error, error) {
	h, err := twitch.Frontend.Helix()
	if err != nil {
		return nil, nil, err
	}

	subs, err := h.ListSubscriptions()
	if err != nil {
		return nil, nil, err
	}

	var fmted []string

	for _, sub := range subs {
		fmted = append(fmted, sub.Type+"="+sub.ID)
	}

	return strings.Join(fmted, " | "), nil, nil
}

/////////////////////
//                 //
// eventsub delete //
//                 //
/////////////////////

var AdminEventSubDelete = adminEventSubDelete{}

type adminEventSubDelete struct{}

func (c adminEventSubDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminEventSubDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminEventSubDelete) Names() []string {
	return core.AliasesDelete
}

func (adminEventSubDelete) Description() string {
	return "Delete a subscription."
}

func (adminEventSubDelete) UsageArgs() string {
	return "<subscription-id...>"
}

func (c adminEventSubDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminEventSubDelete) Examples() []string {
	return nil
}

func (adminEventSubDelete) Parent() core.CommandStatic {
	return AdminEventSub
}

func (adminEventSubDelete) Children() core.CommandsStatic {
	return nil
}

func (adminEventSubDelete) Init() error {
	return nil
}

func (adminEventSubDelete) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	h, err := twitch.Frontend.Helix()
	if err != nil {
		return nil, nil, err
	}

	for _, subid := range m.Command.Args {
		if err := h.DeleteSubscription(subid); err != nil {
			log.Debug().Err(err).Str("id", subid).Msg("failed to delete subscription")
			return "Failed to delete subscription with ID: " + subid, errors.New("failed to delete subscription"), nil
		}

	}
	if len(m.Command.Args) == 1 {
		return "Deleted subscription.", nil, nil
	}
	return "Deleted subscriptions.", nil, nil
}
