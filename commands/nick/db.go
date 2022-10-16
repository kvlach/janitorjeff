package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandNickNicknames (
	user INTEGER NOT NULL,
	place INTEGER NOT NULL,
	nick VARCHAR(255) NOT NULL,
	UNIQUE(user, place),
	FOREIGN KEY (user) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func dbUserAdd(user, place int64, nick string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	INSERT INTO CommandNickNicknames(user, place, nick)
	VALUES (?, ?, ?)`, user, place, nick)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Str("nick", nick).
		Msg("added user nick in db")

	return err
}

func dbUserUpdate(user, place int64, nick string) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE CommandNickNicknames
		SET nick = ?
		WHERE user = ? and place = ?
	`, nick, user, place)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Str("nick", nick).
		Msg("update user nick")

	return err
}

func dbUserExists(user, place int64) (bool, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandNickNicknames
			WHERE user = ? and place = ?
			LIMIT 1
		)`, user, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if scope exists")

	return exists, err
}

func dbUserDelete(user, place int64) error {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandNickNicknames
		WHERE user = ? and place = ?`, user, place)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Msg("deleted user from db")

	return err
}

func dbUserNick(user, place int64) (string, error) {
	db := core.Globals.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var nick string

	row := db.DB.QueryRow(`
		SELECT nick
		FROM CommandNickNicknames
		WHERE user = ? and place = ?`, user, place)

	err := row.Scan(&nick)

	log.Debug().
		Err(err).
		Int64("user", user).
		Int64("place", place).
		Str("nick", nick).
		Msg("got nick for user")

	return nick, err
}
