package twitch

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/twitch"
	"strings"
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
	}
}

func (adminEventSub) Init() error {
	return nil
}

func (adminEventSub) Run(m *core.Message) (any, error, error) {
	return m.Usage(), nil, nil
}

//////////
//      //
// list //
//      //
//////////

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

func (adminEventSubList) Run(m *core.Message) (any, error, error) {
	h, err := m.Client.(*twitch.Twitch).HelixApp()
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