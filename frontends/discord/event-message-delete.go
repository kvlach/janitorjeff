package discord

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func messageDelete(s *dg.Session, m *dg.MessageDelete) {
	rdbKey := rdbMessageReplyToKeyPrefix + m.ID
	if r, err := core.RDB.Get(ctx, rdbKey).Result(); err == nil {
		s.ChannelMessageDelete(m.ChannelID, r)
		if err = core.RDB.Del(ctx, rdbKey).Err(); err != nil {
			log.Debug().Str("key", rdbKey).Msg("failed to delete redis key")
		}
	}
}
