package command

import (
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

const dbShema = `
CREATE TABLE IF NOT EXISTS CommandCommandCommands (
	id INTEGER PRIMARY KEY AUTOINCREMENT,

	place INTEGER NOT NULL,
	trigger VARCHAR(255) NOT NULL,
	response VARCHAR(255) NOT NULL,
	active BOOLEAN NOT NULL,

	creator INTEGER NOT NULL,
	created INTEGER NOT NULL,
	deleter INTEGER,
	deleted INTEGER,

	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (creator) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (deleter) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func _dbAdd(place, creator, timestamp int64, trigger, response string) error {
	db := core.DB

	_, err := db.DB.Exec(`
		INSERT INTO CommandCommandCommands(
			place, trigger, response, active, creator, created
		)
		VALUES (?, ?, ?, ?, ?, ?)`,
		place, trigger, response, true, creator, timestamp)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Str("response", response).
		Int64("creator", creator).
		Int64("timestamp", timestamp).
		Msg("added command")

	return err
}

func dbAdd(place, creator int64, trigger, response string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	timestamp := time.Now().UTC().UnixNano()
	return _dbAdd(place, creator, timestamp, trigger, response)
}

func _dbDel(place, deleter, timestamp int64, trigger string) error {
	db := core.DB

	_, err := db.DB.Exec(`
		UPDATE CommandCommandCommands
		SET active = ?, deleter = ?, deleted = ?
		WHERE place = ? and trigger = ? and active = ?
	`, false, deleter, timestamp, place, trigger, true)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Int64("deleter", deleter).
		Int64("timestamp", timestamp).
		Msg("set command as deleted")

	return err
}

func dbDelete(place, deleter int64, trigger string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	timestamp := time.Now().UTC().UnixNano()
	return _dbDel(place, deleter, timestamp, trigger)
}

func dbEdit(place, editor int64, trigger, response string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	timestamp := time.Now().UTC().UnixNano()

	err := _dbDel(place, editor, timestamp, trigger)
	if err != nil {
		return err
	}

	err = _dbAdd(place, editor, timestamp, trigger, response)
	if err != nil {
		return err
	}

	log.Debug().
		Int64("place", place).
		Int64("author", editor).
		Str("trigger", trigger).
		Str("response", response).
		Msg("changed trigger's response")

	return nil
}

func dbTriggerExists(place int64, trigger string) (bool, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandCommandCommands
			WHERE trigger = ? and place = ? and active = ?
			LIMIT 1
		)`, trigger, place, true)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Str("trigger", trigger).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if trigger exists")

	return exists, err
}

func dbList(place int64) ([]string, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT trigger
		FROM CommandCommandCommands
		WHERE place = ? and active = ?
	`, place, true)
	if err != nil {
		log.Debug().Err(err).Msg("failed to make query")
		return nil, err
	}

	defer rows.Close()

	var triggers []string
	for rows.Next() {
		var trigger string
		if err := rows.Scan(&trigger); err != nil {
			log.Debug().Err(err).Msg("failed while scanning rows")
			return nil, err
		}
		triggers = append(triggers, trigger)
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("place", place).
		Strs("triggers", triggers).
		Msg("got triggers")

	return triggers, err
}

func dbGetResponse(place int64, trigger string) (string, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	row := db.DB.QueryRow(`
		SELECT response
		FROM CommandCommandCommands
		WHERE place = ? and trigger = ? and active = ?
	`, place, trigger, true)

	var response string
	err := row.Scan(&response)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Str("response", response).
		Msg("got response")

	return response, err
}

type customCommand struct {
	response string
	creator  int64
	created  int64
	deleter  int64
	deleted  int64
}

func _dbHistory(place int64, trigger string, active bool) ([]customCommand, error) {
	db := core.DB

	rows, err := db.DB.Query(`
		SELECT response, creator, created, deleter, deleted
		FROM CommandCommandCommands
		WHERE place = ? and trigger = ? and active = ?
	`, place, trigger, active)
	if err != nil {
		log.Debug().Err(err).Msg("failed to make query")
		return nil, err
	}

	defer rows.Close()

	var history []customCommand

	for rows.Next() {
		var response string
		var creator, created, deleter, deleted int64

		var _deleter, _deleted any

		if active == true {
			err = rows.Scan(&response, &creator, &created, &_deleter, &_deleted)
		} else {
			err = rows.Scan(&response, &creator, &created, &deleter, &deleted)
		}

		if err != nil {
			log.Debug().Err(err).Msg("failed while scanning rows")
			return nil, err
		}

		cc := customCommand{
			response: response,
			creator:  creator,
			created:  created,
			deleter:  deleter,
			deleted:  deleted,
		}

		history = append(history, cc)
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("trigger", trigger).
		Interface("history", history).
		Msg("got a command's history")

	return history, err
}

func dbHistory(place int64, trigger string) ([]customCommand, error) {
	db := core.DB
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	inactive, err := _dbHistory(place, trigger, false)
	if err != nil {
		return nil, err
	}

	active, err := _dbHistory(place, trigger, true)
	if err != nil {
		return nil, err
	}

	return append(inactive, active...), nil
}
