package discord

import (
	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(m *core.EventMessage) bool {
	return m.Frontend.Type() == discord.Frontend.Type()
}

func (admin) Names() []string {
	return []string{
		"discord",
	}
}

func (admin) Description() string {
	return "Discord related bot admin operations."
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
		AdminGuild,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

///////////
//       //
// guild //
//       //
///////////

var AdminGuild = adminGuild{}

type adminGuild struct{}

func (c adminGuild) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminGuild) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminGuild) Names() []string {
	return []string{
		"guild",
	}
}

func (adminGuild) Description() string {
	return "Guild related operation commands."
}

func (c adminGuild) UsageArgs() string {
	return c.Children().Usage()
}

func (c adminGuild) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminGuild) Examples() []string {
	return nil
}

func (adminGuild) Parent() core.CommandStatic {
	return Admin
}

func (adminGuild) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminGuildLeave,
	}
}

func (adminGuild) Init() error {
	return nil
}

func (adminGuild) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
}

/////////////////
//             //
// guild leave //
//             //
/////////////////

var AdminGuildLeave = adminGuildLeave{}

type adminGuildLeave struct{}

func (c adminGuildLeave) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminGuildLeave) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminGuildLeave) Names() []string {
	return []string{
		"leave",
		"exit",
	}
}

func (adminGuildLeave) Description() string {
	return "Leave a Discord guild."
}

func (adminGuildLeave) UsageArgs() string {
	return "<guild-id>"
}

func (c adminGuildLeave) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminGuildLeave) Examples() []string {
	return nil
}

func (adminGuildLeave) Parent() core.CommandStatic {
	return AdminGuild
}

func (adminGuildLeave) Children() core.CommandsStatic {
	return nil
}

func (adminGuildLeave) Init() error {
	return nil
}

func (c adminGuildLeave) Run(m *core.EventMessage) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	err := discord.Client.Session.GuildLeave(m.Command.Args[0])
	embed := &dg.MessageEmbed{
		Description: c.fmt(err),
	}
	return embed, err, nil
}

func (adminGuildLeave) fmt(err error) string {
	switch err {
	case nil:
		return "Left guild."
	default:
		return "Failed to leave guild: " + err.Error()
	}
}
