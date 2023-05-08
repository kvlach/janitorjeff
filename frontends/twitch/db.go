package twitch

import (
	"database/sql"
	"git.sr.ht/~slowtyper/janitorjeff/core"
)

func dbAddChannelSimple(uid, uname string) (int64, error) {
	return dbAddChannel(uid, uid, uname, nil)
}

func dbAddChannel(id string, uid, uname string, h *Helix) (int64, error) {
	var channelID, channelName string
	if id == uid {
		channelID = uid
		channelName = uname
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
		INSERT INTO frontend_twitch_channels(scope, channel_id, channel_name)
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
		SELECT scope
		FROM frontend_twitch_channels
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
		FROM frontend_twitch_channels
		WHERE scope = $1`, scope)

	var id, name string
	err := row.Scan(&id, &name)
	return id, name, err
}

func dbSetUserAccessToken(scope int64, accessToken, refreshToken string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE frontend_twitch_channels
		SET access_token = $1, refresh_token = $2
		WHERE scope = $3`, accessToken, refreshToken, scope)

	return err
}

func dbUpdateUserTokens(oldAcessToken, accessToken, refreshToken string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE frontend_twitch_channels
		SET access_token = $1, refresh_token = $2
		WHERE access_token = $3`, accessToken, refreshToken, oldAcessToken)

	return err
}

func dbGetUserAccessToken(channelID string) (string, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	row := db.DB.QueryRow("SELECT access_token FROM frontend_twitch_channels WHERE channel_id = $1", channelID)

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
		FROM frontend_twitch_channels
		WHERE access_token = $1`, accessToken)

	var refreshToken string
	err := row.Scan(&refreshToken)
	return refreshToken, err
}

func AddEventSubSubscriptionID(tx *sql.Tx, id string) error {
	_, err := tx.Exec(`
		INSERT INTO frontend_twitch_eventsub(id)
		VALUES ($1)`, id)
	return err
}

func DeleteEventSubSubscriptionID(tx *sql.Tx, id string) error {
	_, err := tx.Exec(`
		DELETE FROM frontend_twitch_eventsub
		WHERE id = $1`, id)
	return err
}
