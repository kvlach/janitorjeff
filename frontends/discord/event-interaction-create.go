package discord

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type InteractionCreate struct {
	Interaction *dg.InteractionCreate
	Data        *dg.ApplicationCommandInteractionData
	VC          *dg.VoiceConnection
}

func RegisterAppCommand(cmd *dg.ApplicationCommand) {
	guildID := "759669782386966528"

	cmd, err := Client.Session.ApplicationCommandCreate(Client.Session.State.User.ID, guildID, cmd)
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

	inter := &InteractionCreate{
		Interaction: i,
		Data:        &data,
	}

	m, err := inter.Parse()
	if err != nil {
		log.Debug().Err(err).Send()
		return
	}

	var prefix string
	for _, p := range core.Prefixes.Others() {
		if p.Type == core.Normal {
			prefix = p.Prefix
			break
		}
	}

	cmd, index, _ := core.Commands.Match(Type, m, args)

	m.Command = &core.Command{
		CommandStatic: cmd,
		CommandRuntime: core.CommandRuntime{
			Path:   args[:index+1],
			Args:   args[index+1:],
			Prefix: prefix,
		},
	}

	m.Raw = prefix + strings.Join(args, " ")
	fmt.Println("MESSAGEEEEEEEEEEEEEEEEEEEEE", m.Raw)

	resp, urr, err := cmd.Run(m)
	if err == core.UrrSilence {
		return
	}
	if err != nil {
		m.Write("Something went wrong...", fmt.Errorf(""))
		return
	}
	m.Write(resp, urr)
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (i *InteractionCreate) Parse() (*core.Message, error) {
	author := &AuthorInteraction{
		GuildID: i.Interaction.GuildID,
		Member:  i.Interaction.Member,
		User:    i.Interaction.User,
	}

	h := &Here{
		ChannelID: i.Interaction.ChannelID,
		GuildID:   i.Interaction.GuildID,
		Author:    author,
	}

	sp := &Speaker{
		Author: author,
		Here:   h,
		VC:     nil,
	}

	m := &core.Message{
		ID:       i.Data.ID,
		Raw:      "", // TODO
		Frontend: Frontend,
		Author:   author,
		Here:     h,
		Client:   i,
		Speaker:  sp,
	}
	return m, nil
}

func (i *InteractionCreate) PersonID(s, placeID string) (string, error) {
	var id string
	if i.Interaction.Member != nil {
		id = i.Interaction.Member.User.ID
	} else {
		id = i.Interaction.User.ID
	}
	return getPersonID(s, placeID, id)
}

func (i *InteractionCreate) PlaceID(s string) (string, error) {
	return getPlaceID(s)
}

func (i *InteractionCreate) Person(id string) (int64, error) {
	return dbGetPersonScope(id)
}

func (i *InteractionCreate) send(msg any, urr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		resp := &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Content: msg.(string),
			},
		}
		return nil, Client.Session.InteractionRespond(i.Interaction.Interaction, resp)

	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		embed = embedColor(embed, urr)

		resp := &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Embeds: []*dg.MessageEmbed{
					embed,
				},
			},
		}
		return nil, Client.Session.InteractionRespond(i.Interaction.Interaction, resp)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (i *InteractionCreate) Send(msg any, urr core.Urr) (*core.Message, error) {
	return i.send(msg, urr)
}

func (i *InteractionCreate) Ping(msg any, urr core.Urr) (*core.Message, error) {
	return i.send(msg, urr)
}

func (i *InteractionCreate) Write(msg any, urr core.Urr) (*core.Message, error) {
	return i.send(msg, urr)
}

func (i *InteractionCreate) Natural(msg any, urr core.Urr) (*core.Message, error) {
	return i.send(msg, urr)
}

func (i *InteractionCreate) QuoteCommand(cmd string) string {
	return PlaceInBackticks(cmd)
}

/////////////
//         //
// Speaker //
//         //
/////////////

func (i *InteractionCreate) Enabled() bool {
	return true
}

func (i *InteractionCreate) FrameRate() int {
	return frameRate
}

func (i *InteractionCreate) Channels() int {
	return channels
}

func (i *InteractionCreate) Join() error {
	var userID string
	if i.Interaction.Member != nil {
		userID = i.Interaction.Member.User.ID
	} else {
		userID = i.Interaction.User.ID
	}

	v, err := Client.VoiceJoin(i.Interaction.GuildID, userID)
	if err != nil {
		return err
	}
	i.VC = v
	return nil
}

func (i *InteractionCreate) Leave() error {
	if i.VC == nil {
		return errors.New("not connected, can't disconnect")
	}
	if err := i.VC.Disconnect(); err != nil {
		return err
	}
	i.VC = nil
	return nil
}

func (i *InteractionCreate) Say(buf io.Reader, s <-chan core.AudioState) error {
	return voicePlay(i.VC, buf, s)
}

func (i *InteractionCreate) AuthorDeafened() (bool, error) {
	var authorID string

	if i.Interaction.Member != nil {
		authorID = i.Interaction.Message.Author.ID
	} else {
		authorID = i.Interaction.User.ID
	}

	vs, err := Client.VoiceState(i.Interaction.GuildID, authorID)
	if err != nil {
		return false, err
	}
	return vs.SelfDeaf, nil
}

func (i *InteractionCreate) AuthorConnected() (bool, error) {
	// TODO: implement this
	return false, nil
}
