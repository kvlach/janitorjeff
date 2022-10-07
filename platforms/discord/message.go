package discord

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

type DiscordMessage struct {
	session *dg.Session
	message *dg.Message
}

func (d *DiscordMessage) Parse() (*core.Message, error) {
	author := &core.Author{
		ID:          d.message.Author.ID,
		Name:        d.message.Author.Username,
		DisplayName: getDisplayName(d.message.Member, d.message.Author),
		Mention:     d.message.Author.Mention(),
	}

	channel := &core.Channel{
		ID:   d.message.ChannelID,
		Name: d.message.ChannelID,
	}

	msg := &core.Message{
		ID:   d.message.ID,
		Type: core.Discord,
		Raw:  d.message.Content,
		// GuildID is always empty in returned message objects, this is here in
		// case that changes in the future.
		IsDM:    d.message.GuildID == "",
		Author:  author,
		Channel: channel,
		Client:  d,
	}

	return msg, nil
}

func (d *DiscordMessage) Scope(type_ int) (int64, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	switch type_ {
	case Default, Guild, Channel, Thread:
		return getScopePlace(type_, d.message.ChannelID, d.message.GuildID)
	case Author:
		return getScopeAuthor(d.message.Author.ID)
	default:
		return -1, fmt.Errorf("type '%d' not supported", type_)
	}
}

func (d *DiscordMessage) Write(msg interface{}, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		text := msg.(string)
		lenLim := 2000
		lenCnt := func(s string) int { return len(s) }
		return messagesTextSend(d.session, text, d.message.ChannelID, lenLim, lenCnt)

	case *dg.MessageEmbed:
		// TODO
		return nil, nil
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}
