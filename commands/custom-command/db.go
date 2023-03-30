package custom_command

import (
	"time"

	"github.com/janitorjeff/jeff-bot/core"

	"github.com/rs/zerolog/log"
)

func _dbAdd(place, creator, timestamp int64, trigger, response string) error {
	db := core.DB

	_, err := db.DB.Exec(`
		INSERT INTO cmd_customcommand_commands(
			place, trigger, response, active, creator, created
		)
		VALUES ($1, $2, $3, $4, $5, $6)`,
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

	timestamp := time.Now().UTC().Unix()
	return _dbAdd(place, creator, timestamp, trigger, response)
}

func _dbDel(place, deleter, timestamp int64, trigger string) error {
	db := core.DB

	_, err := db.DB.Exec(`
		UPDATE cmd_customcommand_commands
		SET active = $1, deleter = $2, deleted = $3
		WHERE place = $4 and trigger = $5 and active = $6
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

	timestamp := time.Now().UTC().Unix()
	return _dbDel(place, deleter, timestamp, trigger)
}

func dbEdit(place, editor int64, trigger, response string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	timestamp := time.Now().UTC().Unix()

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
			SELECT 1 FROM cmd_customcommand_commands
			WHERE trigger = $1 and place = $2 and active = $3
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
		FROM cmd_customcommand_commands
		WHERE place = $1 and active = $2
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
		FROM cmd_customcommand_commands
		WHERE place = $1 and trigger = $2 and active = $3
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
		FROM cmd_customcommand_commands
		WHERE place = $1 and trigger = $2 and active = $3
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
