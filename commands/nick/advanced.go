package nick

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"

	dg "github.com/bwmarrin/discordgo"
)

var Advanced = &core.CommandStatic{
	Names: []string{
		"nick",
		"nickname",
	},
	Description: "Set, view or delete your nickname.",
	UsageArgs:   "(view|set|delete)",
	Run:         advancedRun,
	Init:        init_,

	Children: core.Commands{
		{
			Names: []string{
				"view",
				"get",
			},
			Description: "View your current nickname.",
			UsageArgs:   "",
			Run:         advancedRunView,
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
			Names: []string{
				"delete",
				"del",
				"remove",
				"rm",
			},
			Description: "Delete your nickname.",
			UsageArgs:   "",
			Run:         advancedRunDelete,
		},
	},
}

func advancedRun(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}

//////////
//      //
// view //
//      //
//////////

func advancedRunView(m *core.Message) (any, error, error) {
	switch m.Type {
	case frontends.Discord:
		return advancedRunViewDiscord(m)
	default:
		return advancedRunViewText(m)
	}
}

func advancedRunViewDiscord(m *core.Message) (*dg.MessageEmbed, error, error) {
	nick, usrErr, err := advancedRunViewCore(m)
	if err != nil {
		return nil, nil, err
	}

	nick = fmt.Sprintf("**%s**", nick)

	embed := &dg.MessageEmbed{
		Description: advancedRunViewErr(usrErr, nick),
	}

	return embed, usrErr, nil
}

func advancedRunViewText(m *core.Message) (string, error, error) {
	nick, usrErr, err := advancedRunViewCore(m)
	if err != nil {
		return "", nil, err
	}
	nick = fmt.Sprintf("'%s'", nick)
	return advancedRunViewErr(usrErr, nick), usrErr, nil
}

func advancedRunViewErr(usrErr error, nick string) string {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Your nickname is: %s", nick)
	case errPersonNotFound:
		return "You have not set a nickname."
	default:
		return fmt.Sprint(usrErr)
	}
}

func advancedRunViewCore(m *core.Message) (string, error, error) {
	author, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	return runGet(author, place)
}

/////////
//     //
// set //
//     //
/////////

func advancedRunSet(m *core.Message) (any, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return m.ReplyUsage(), core.ErrMissingArgs, nil
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

	author, err := m.ScopeAuthor()
	if err != nil {
		return "", nil, err
	}

	place, err := m.HereLogical()
	if err != nil {
		return "", nil, err
	}

	usrErr, err := runSet(nick, author, place)
	return nick, usrErr, err
}

////////////
//        //
// delete //
//        //
////////////

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
	author, err := m.ScopeAuthor()
	if err != nil {
		return nil, err
	}

	place, err := m.HereLogical()
	if err != nil {
		return nil, err
	}

	return runDelete(author, place)
}
