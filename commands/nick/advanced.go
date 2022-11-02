package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"

	dg "github.com/bwmarrin/discordgo"
)

///////////////////
//               //
// Type: Command //
//               //
///////////////////

var Advanced = &core.CommandStatic{
	Names: []string{
		"nick",
		"nickname",
	},
	Description: "Set, view or delete your nickname.",
	UsageArgs:   "(show | set | delete)",
	Run:         advancedRun,
	Init:        advancedInit,

	Children: core.Commands{
		{
			Names:       core.Show,
			Description: "View your current nickname.",
			UsageArgs:   "",
			Run:         advancedRunShow,
		},
		{
			Names: []string{
				"set",
			},
			Description: "Set your nickname.",
			UsageArgs:   "<nickname>",
			Run:         advancedRunSet,
		},
		{
			Names:       core.Delete,
			Description: "Delete your nickname.",
			UsageArgs:   "",
			Run:         advancedRunDelete,
		},
	},
}

///////////////
//           //
// Type: Run //
//           //
///////////////

func advancedRun(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

///////////////
//           //
// Run: show //
//           //
///////////////

func advancedRunShow(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunShowDiscord(m)
	default:
		return advancedRunShowText(m)
	}
}

func advancedRunShowDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := advancedRunShowCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: advancedRunShowErr(usrErr, nick),
	}

	return embed, usrErr, nil
}

func advancedRunShowText(m *core.Message) (string, error, error) {
	nick, usrErr, err := advancedRunShowCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return advancedRunShowErr(usrErr, nick), usrErr, nil
}

func advancedRunShowErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your nickname is: %s", nick)
	case errPersonNotFound:
		return "You have not set a nickname."
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunShowCore(m *core.Message) (string, error, error) {
	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	return runShow(author, here)
}

//////////////
//          //
// Run: set //
//          //
//////////////

func advancedRunSet(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Type {
	case frontends.Discord:
		return advancedRunSetDiscord(m)
	default:
		return advancedRunSetText(m)
	}
}

func advancedRunSetDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := advancedRunSetCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: advancedRunSetErr(usrErr, nick),
	}

	return embed, usrErr, nil
}

func advancedRunSetText(m *core.Message) (string, error, error) {
	nick, usrErr, err := advancedRunSetCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return advancedRunSetErr(usrErr, nick), usrErr, nil
}

func advancedRunSetErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Nickname set to %s", nick)
	case errNickExists:
		return fmt.Sprintf("Nickname %s is already being used by another user.", nick)
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunSetCore(m *core.Message) (string, error, error) {
	nick := m.Command.Runtime.Args[0]

	author, err := m.Author()
	if err != nil {
		return "", nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	usrErr, err := runSet(nick, author, here)
	return nick, usrErr, err
}

/////////////////
//             //
// Run: delete //
//             //
/////////////////

func advancedRunDelete(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunDeleteDiscord(m)
	default:
		return advancedRunDeleteText(m)
	}
}

func advancedRunDeleteDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	usrErr, err := advancedRunDeleteCore(m)
	if err != nil {
		return nil, nil, err
	}

	embed := &dg.MessageEmbed{
		Description: advancedRunDeleteErr(usrErr),
	}

	return embed, usrErr, nil
}

func advancedRunDeleteText(m *core.Message) (string, error, error) {
	usrErr, err := advancedRunDeleteCore(m)
	if err != nil {
		return "", nil, err
	}
	return advancedRunDeleteErr(usrErr), usrErr, nil
}

func advancedRunDeleteErr(usrErr error) string {
	switch usrErr {
	case nil:
		return "Deleted your nickname."
	case errPersonNotFound:
		return "Can't delete, you haven't set a nickname here."
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunDeleteCore(m *core.Message) (error, error) {
	author, err := m.Author()
	if err != nil {
		return nil, err
	}

	here, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	return runDelete(author, here)
}

////////////////
//            //
// Type: Init //
//            //
////////////////

func advancedInit() error {
	advancedInitDiscordAppCommand()
	return nil
}

func advancedInitDiscordAppCommand() {
	cmd := &dg.ApplicationCommand{
		Name:        "nick",
		Type:        dg.ChatApplicationCommand,
		Description: "Nickname commands.",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "show",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: "Show your nickname.",
			},
			{
				Name:        "set",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: "Set your nickname.",
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "nickname",
						Type:        dg.ApplicationCommandOptionString,
						Description: "give nickname",
						Required:    true,
					},
				},
			},
			{
				Name:        "delete",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Description: "Delete your nickname.",
			},
		},
	}

	discord.RegisterAppCommand(cmd)
}
