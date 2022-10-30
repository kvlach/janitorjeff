package discord

import (
	dg "github.com/bwmarrin/discordgo"
)

func messageDelete(s *dg.Session, m *dg.MessageDelete) {
	if r, ok := replies.Get(m.ID); ok {
		s.ChannelMessageDelete(m.ChannelID, r)
		replies.Delete(r)
	}
}
