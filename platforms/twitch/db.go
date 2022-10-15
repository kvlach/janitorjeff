package twitch

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	twitchIRC "github.com/gempir/go-twitch-irc/v2"
)

func twitchChannelAddChannel(t int, id string, msg *twitchIRC.PrivateMessage) (int64, error) {
	var channelID string
	var channelName string

	switch t {
	case Default, Author:
		channelID = msg.User.ID
		channelName = msg.User.Name
	case Channel, User:
		channelID = id
		channelName = "" // TODO: fix this?
	default:
		return -1, fmt.Errorf("type '%d' not supproted", t)
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

	scope, err = db.ScopeAdd(tx)
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

	_, err := twitchChannelAddChannel(Channel, channelID, channelName)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(`
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
