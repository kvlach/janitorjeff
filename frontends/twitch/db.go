package twitch

import (
	"github.com/janitorjeff/jeff-bot/core"

	tirc "github.com/gempir/go-twitch-irc/v2"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS PlatformTwitchChannels (
	id INTEGER PRIMARY KEY,
	channel_id VARCHAR(255) NOT NULL UNIQUE,
	channel_name VARCHAR(255) NOT NULL,
	access_token VARCHAR(255),
	refresh_token VARCHAR(255),
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);
`

func dbInit() error {
	return core.DB.Init(dbSchema)
}

func dbAddChannelSimple(uid, uname string) (int64, error) {
	u := tirc.User{
		ID:   uid,
		Name: uname,
	}
	return dbAddChannel(uid, u, nil)
}

func dbAddChannel(id string, u tirc.User, h *Helix) (int64, error) {
	var channelID, channelName string
	if id == u.ID {
		channelID = u.ID
		channelName = u.Name
	} else if u, err := h.GetUser(id); err == nil {
		channelID = id
		channelName = u.Login
	} else {
		return -1, err
	}

	// if scope exists return it instead of re-adding it
	scope, err := dbGetChannelScope(channelID)
	if err == nil {
		return scope, nil
	}

	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	scope, err = db.ScopeAdd(tx, channelID, Type)
	if err != nil {
		return -1, err
	}

	_, err = tx.Exec(`
		INSERT INTO PlatformTwitchChannels(id, channel_id, channel_name)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;`, scope, channelID, channelName)

	if err != nil {
		return -1, err
	}

	return scope, tx.Commit()
}

func dbGetChannelScope(channelID string) (int64, error) {
	db := core.DB

	row := db.DB.QueryRow(`
		SELECT id
		FROM PlatformTwitchChannels
		WHERE channel_id = $1`, channelID)

	var id int64
	err := row.Scan(&id)
	return id, err
}

func dbGetChannel(scope int64) (string, string, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	row := db.DB.QueryRow(`
		SELECT channel_id, channel_name
		FROM PlatformTwitchChannels
		WHERE id = $1`, scope)

	var id, name string
	err := row.Scan(&id, &name)
	return id, name, err
}

func dbSetUserAccessToken(accessToken, refreshToken, channelID string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	// _, err := twitchChannelAddChannel(Channel, channelID, channelID, channelName)
	// if err != nil {
	// 	return err
	// }

	_, err := db.DB.Exec(`
		UPDATE PlatformTwitchChannels
		SET access_token = $1, refresh_token = $2
		WHERE channel_id = $3`, accessToken, refreshToken, channelID)

	return err
}

func dbUpdateUserTokens(oldAcessToken, accessToken, refreshToken string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE PlatformTwitchChannels
		SET access_token = $1, refresh_token = $2
		WHERE access_token = $3`, accessToken, refreshToken, oldAcessToken)

	return err
}

func dbGetUserAccessToken(channelID string) (string, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	row := db.DB.QueryRow("SELECT access_token FROM PlatformTwitchChannels WHERE channel_id = $1", channelID)

	var accessToken string
	err := row.Scan(&accessToken)
	return accessToken, err
}

func dbGetetUserRefreshToken(accessToken string) (string, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	row := db.DB.QueryRow(`
		SELECT refresh_token
		FROM PlatformTwitchChannels
		WHERE access_token = $1`, accessToken)

	var refreshToken string
	err := row.Scan(&refreshToken)
	return refreshToken, err
}
