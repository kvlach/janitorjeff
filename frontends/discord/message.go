package discord

import (
	"fmt"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
)

type Message struct {
	Message *dg.Message
	VC      *dg.VoiceConnection
}

///////////////
//           //
// Messenger //
//           //
///////////////

func (d *Message) Parse() (*core.Message, error) {
	msg := parse(d.Message)
	msg.Client = d
	return msg, nil
}

func (d *Message) PersonID(s, placeID string) (string, error) {
	return getPersonID(s, placeID, d.Message.Author.ID)
}

func (d *Message) PlaceID(s string) (string, error) {
	return getPlaceID(s)
}

func (d *Message) Person(id string) (int64, error) {
	return dbGetPersonScope(id)
}

func (d *Message) send(msg any, urr error, ping bool) (*core.Message, error) {
	switch t := msg.(type) {
	case string:
		return sendText(d.Message, msg.(string), ping)
	case *dg.MessageEmbed:
		embed := msg.(*dg.MessageEmbed)
		return sendEmbed(d.Message, embed, urr, ping)
	default:
		return nil, fmt.Errorf("Can't send discord message of type %v", t)
	}
}

func (d *Message) Send(msg any, urr core.Urr) (*core.Message, error) {
	return d.send(msg, urr, false)
}

func (d *Message) Ping(msg any, urr core.Urr) (*core.Message, error) {
	return d.send(msg, urr, true)
}

func (d *Message) Write(msg any, urr core.Urr) (*core.Message, error) {
	return d.Send(msg, urr)
}

func (d *Message) Natural(msg any, urr core.Urr) (*core.Message, error) {
	return d.Send(msg, urr)
}

func (d *Message) QuoteCommand(cmd string) string {
	return PlaceInBackticks(cmd)
}
