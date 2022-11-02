package discord

import (
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type InteractionCreate struct {
	Session     *dg.Session
	Interaction *dg.InteractionCreate
	Data        *dg.ApplicationCommandInteractionData
}

func getUserID(i *InteractionCreate) string {
	if i.Interaction.Member != nil {
		return i.Interaction.Member.User.ID
	}
	return i.Interaction.User.ID
}

func (i *InteractionCreate) Admin() bool {
	return isAdmin(getUserID(i))
}

func (i *InteractionCreate) Parse() (*core.Message, error) {
	var user *core.User
	if i.Interaction.Member != nil {
		user = &core.User{
			ID:          i.Interaction.Member.User.ID,
			Name:        i.Interaction.Member.User.Username,
			DisplayName: i.Interaction.Member.User.Username,
			Mention:     i.Interaction.Member.Mention(),
		}
	} else {
		user = &core.User{
			ID:          i.Interaction.User.ID,
			Name:        i.Interaction.User.Username,
			DisplayName: i.Interaction.User.Username,
			Mention:     i.Interaction.User.Mention(),
		}
	}

	channel := &core.Channel{
		ID:   i.Interaction.ChannelID,
		Name: i.Interaction.ChannelID,
	}

	m := &core.Message{
		ID:      i.Data.ID,
		Type:    Type,
		Raw:     "", // TODO
		IsDM:    i.Interaction.User != nil,
		User:    user,
		Channel: channel,
		Client:  i,
	}
	return m, nil
}

func (i *InteractionCreate) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, getUserID(i), i.Session)
}

func (i *InteractionCreate) PlaceID(s string) (string, error) {
	return getPlaceID(s, i.Session)
}

func (i *InteractionCreate) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (i *InteractionCreate) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, i.Interaction.ChannelID, i.Interaction.GuildID, i.Session)
}

func (i *InteractionCreate) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, i.Interaction.ChannelID, i.Interaction.GuildID, i.Session)
}

func (i *InteractionCreate) Usage(usage string) any {
	return getUsage(usage)
}

func (i *InteractionCreate) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(i.Session, msg.(string), i.Interaction.ChannelID, i.Interaction.GuildID)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		embed = embedColor(embed, usrErr)

		resp := &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					embed,
				},
			},
		}
		return nil, i.Session.InteractionRespond(i.Interaction.Interaction, resp)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func RegisterAppCommand(cmd *dg.ApplicationCommand) {
	guildID := "759669782386966528"

	s := core.Globals.Discord.Client

	cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
	if err != nil {
		panic(err)
	}
	fmt.Println(cmd)
}

func interactionCreate(s *dg.Session, i *dg.InteractionCreate) {
	if i.Type != dg.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()

	args := []string{data.Name}

	opts := data.Options
	for len(opts) != 0 && opts[0].Type == dg.ApplicationCommandOptionSubCommand {
		args = append(args, opts[0].Name)
		opts = opts[0].Options
	}

	if len(opts) != 0 {
		if val := opts[0].Value; val != nil {
			args = append(args, strings.Split(fmt.Sprint(val), " ")...)
		}
	}

	cmd, index, _ := core.Globals.Commands.Advanced.MatchCommand(args)
	inter := &InteractionCreate{
		Session:     s,
		Interaction: i,
		Data:        &data,
	}

	m, err := inter.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	var prefix string
	for _, p := range core.Globals.Prefixes.Others {
		if p.Type == core.Normal {
			prefix = p.Prefix
			break
		}
	}

	m.Command = &core.Command{
		Type:   core.Advanced,
		Static: cmd,
		Runtime: &core.CommandRuntime{
			Name:   args[:index+1],
			Args:   args[index+1:],
			Prefix: prefix,
		},
	}

	m.Raw = prefix + strings.Join(args, " ")
	fmt.Println("MESSAGEEEEEEEEEEEEEEEEEEEEE", m.Raw)

	resp, usrErr, err := cmd.Run(m)
	if err == core.ErrSilence {
		return
	}
	if err != nil {
		m.Write("Something went wrong...", fmt.Errorf(""))
		return
	}
	m.Write(resp, usrErr)
}
