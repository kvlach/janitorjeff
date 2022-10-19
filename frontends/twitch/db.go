package twitch

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	tirc "github.com/gempir/go-twitch-irc/v2"
)

func twitchChannelAddChannel(id string, m *tirc.PrivateMessage, h *Helix) (int64, error) {
	var channelID, channelName string
	if id == m.User.ID {
		channelID = m.User.ID
		channelName = m.User.Name
	} else if name, err := h.GetUserName(id); err == nil {
		channelID = id
		channelName = name
	} else {
		return -1, err
	}

	// if scope exists return it instead of re-adding it
	scope, err := twitchGetChannelScope(channelID)
	if err == nil {
		return scope, nil
	}

	db := core.Globals.DB

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	scope, err = db.ScopeAdd(tx, channelID)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO PlatformTwitchChannels(id, channel_id, channel_name)
		VALUES (?, ?, ?)`, scope, channelID, channelName)

	if err != nil {
		return -1, err
	}

	return scope, tx.Commit()
}

func twitchGetChannelScope(channelID string) (int64, error) {
	db := core.Globals.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformTwitchChannels
		WHERE channel_id = ?`, channelID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func TwitchChannelSetAccessToken(accessToken, refreshToken, channelID, channelName string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	// _, err := twitchChannelAddChannel(Channel, channelID, channelID, channelName)
	// if err != nil {
	// 	return err
	// }

	_, err := db.DB.Exec(`
		UPDATE PlatformTwitchChannels
		SET access_token = ?, refresh_token = ?
		WHERE channel_id = ?`, accessToken, refreshToken, channelID)

	return err
}

func twitchChannelGetAccessToken(channelID string) (string, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	row := db.DB.QueryRow("SELECT access_token FROM PlatformTwitchChannels WHERE channel_id = ?", channelID)

	var accessToken string
	err := row.Scan(&accessToken)
	return accessToken, err
}
