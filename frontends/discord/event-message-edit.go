package discord

import (
	"fmt"
	"sync"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

var replies = replyCache{}

type replyCache struct {
	lock    sync.RWMutex
	replies map[string]string
}

func (r *replyCache) Set(key, value string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.replies == nil {
		r.replies = map[string]string{}
	}
	r.replies[key] = value
}

func (r *replyCache) Get(key string) (string, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	value, ok := r.replies[key]
	return value, ok
}

type DiscordMessageEdit struct {
	Session *dg.Session
	Message *dg.MessageUpdate
}

func (d *DiscordMessageEdit) Admin() bool {
	return isAdmin(d.Message.Author.ID)
}

func (d *DiscordMessageEdit) Parse() (*core.Message, error) {
	msg := parse(d.Message.Message)
	msg.Client = d
	return msg, nil
}

func (d *DiscordMessageEdit) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Session, d.Message.Message)
}

func (d *DiscordMessageEdit) PlaceID(s string) (string, error) {
	return getPlaceID(s, d.Session, d.Message.Message)
}

func (d *DiscordMessageEdit) Person(id string) (int64, error) {
	return getPersonScope(id)
}

func (d *DiscordMessageEdit) PlaceExact(id string) (int64, error) {
	return getPlaceExactScope(id, d.Message.Message, d.Session)
}

func (d *DiscordMessageEdit) PlaceLogical(id string) (int64, error) {
	return getPlaceLogicalScope(id, d.Message.Message, d.Session)
}

func (d *DiscordMessageEdit) ReplyUsage(usage string) any {
	return replyUsage(usage)
}

func (d *DiscordMessageEdit) Write(msg any, usrErr error) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Session, msg.(string), d.Message.ChannelID, d.Message.GuildID)

	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		id, ok := replies.Get(d.Message.ID)
		if !ok {
			return sendEmbed(d.Session, d.Message.Message, embed, usrErr)
		}
		return editEmbed(d.Session, d.Message.Message, embed, usrErr, id)

	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}

}

func messageEdit(s *dg.Session, m *dg.MessageUpdate) {
	// For some reason this randomly gets triggered if a link has been sent
	// and there has been no edit to the message. Perhaps it could be have
	// something to do with the message getting automatically edited in order
	// for the embed to added. This is difficult to test as it usually doesn't
	// happen.
	if m.Author == nil {
		return
	}

	if m.Author.Bot {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	d := &DiscordMessageEdit{s, m}
	msg, err := d.Parse()
	if err != nil {
		return
	}
	msg.Run()
}
