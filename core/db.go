package core

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var DB *SQLDB

var ctx = context.Background()

type SQLDB struct {
	Lock sync.RWMutex
	DB   *sql.DB
}

func Open(driver, source string) (*SQLDB, error) {
	sqlDB, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	return &SQLDB{DB: sqlDB}, nil
}

func (db *SQLDB) Close() error {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	return db.DB.Close()
}

func (db *SQLDB) Init(schema string) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	return tx.Commit()
}

func (_ *SQLDB) ScopeAdd(tx *sql.Tx, frontendID string, frontend int) (int64, error) {
	var id int64
	err := tx.QueryRow(`
		INSERT INTO scopes(frontend_id, frontend_type)
		VALUES ($1, $2) RETURNING id;`, frontendID, frontend).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// ScopeID returns the given scope's frontend specific ID
func (db *SQLDB) ScopeID(scope int64) (string, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var id string
	row := db.DB.QueryRow(`
		SELECT frontend_id
		FROM scopes
		WHERE id = $1
	`, scope)

	err := row.Scan(&id)

	return id, err
}

// ScopeFrontend returns the given scope's frontend id
func (db *SQLDB) ScopeFrontend(scope int64) (int64, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var id int64
	row := db.DB.QueryRow(`
		SELECT frontend_type
		FROM scopes
		WHERE id = $1
	`, scope)

	err := row.Scan(&id)

	return id, err
}

// PrefixList returns the list of all prefixes for a specific scope.
func (db *SQLDB) PrefixList(place int64) ([]Prefix, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT prefix, type
		FROM prefixes
		WHERE place = $1`, place)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []Prefix
	for rows.Next() {
		var prefix string
		var t CommandType
		if err := rows.Scan(&prefix, &t); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, Prefix{Type: t, Prefix: prefix})
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface("prefixes", prefixes)

	return prefixes, err
}

////////////////////
//                //
// place settings //
//                //
////////////////////

func settingsPlaceExist(tx *sql.Tx, place int64) (bool, error) {
	var exists bool

	row := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM settings_place
			WHERE place = $1
			LIMIT 1
		);
	`, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("place", place).
		Bool("exist", exists).
		Msg("checked if place settings exist")

	return exists, err
}

func settingsPlaceGenerate(tx *sql.Tx, place int64) error {
	_, err := tx.Exec(`
		INSERT INTO settings_place (place)
		VALUES ($1)
	`, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Msg("generated place settings")

	return err
}

// SettingsPlaceGenerate will check if settings for the specified place exist
// and if not will generate them.
func SettingsPlaceGenerate(tx *sql.Tx, place int64) error {
	rdbKey := fmt.Sprintf("settings_place_%d", place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		log.Debug().Msg("CACHE: place settings already generated")
		return nil
	}

	exists, err := settingsPlaceExist(tx, place)
	if err != nil {
		return err
	}
	if exists {
		err := RDB.Set(ctx, rdbKey, nil, 0).Err()
		log.Debug().
			Err(err).
			Msg("CACHE: place settings already exist in db, caching")
		return err
	}

	err = settingsPlaceGenerate(tx, place)
	if err != nil {
		return err
	}
	err = RDB.Set(ctx, rdbKey, nil, 0).Err()
	log.Debug().Err(err).Msg("CACHE: generated place settings in db, caching")
	return err
}

// SettingPlaceGet returns the value of col in table for the specified place.
func (db *SQLDB) SettingPlaceGet(col string, place int64) (any, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return false, err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	// Make sure that the place settings are present
	if err := SettingsPlaceGenerate(tx, place); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM settings_place
		WHERE place = $1
	`, col)

	row := tx.QueryRow(query, place)

	var val any
	err = row.Scan(&val)
	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("got value")
	if err != nil {
		return nil, err
	}
	return val, tx.Commit()
}

// SettingPlaceSet sets the value of col in table for the specified place.
func (db *SQLDB) SettingPlaceSet(col string, place int64, val any) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	// Make sure that the place settings are present
	if err := SettingsPlaceGenerate(tx, place); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE settings_place
		SET %s = $1
		WHERE place = $2
	`, col)
	_, err = tx.Exec(query, val, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("changed setting")

	if err != nil {
		return err
	}
	return tx.Commit()
}

/////////////////////
//                 //
// person settings //
//                 //
/////////////////////

func settingsPersonExist(tx *sql.Tx, person, place int64) (bool, error) {
	var exists bool

	row := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM settings_person
			WHERE person = $1 and place = $2
			LIMIT 1
		)
	`, person, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Bool("exist", exists).
		Msg("checked if person settings exist")

	return exists, err
}

func settingsPersonGenerate(tx *sql.Tx, person, place int64) error {
	_, err := tx.Exec(`
		INSERT INTO settings_person (person, place)
		VALUES ($1, $2)
	`, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Msg("generated person settings")

	return err
}

// SettingsPersonGenerate will check if settings for the specified person in the
// specified place exist, and if not will generate them.
func SettingsPersonGenerate(tx *sql.Tx, person, place int64) error {
	rdbKey := fmt.Sprintf("settings_person_%d_%d", person, place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		log.Debug().Msg("CACHE: person settings already generated")
		return nil
	}

	exists, err := settingsPersonExist(tx, person, place)
	if err != nil {
		return err
	}
	if exists {
		err := RDB.Set(ctx, rdbKey, nil, 0).Err()
		log.Debug().
			Err(err).
			Msg("CACHE: person settings already exist in db, caching")
		return err
	}

	err = settingsPersonGenerate(tx, person, place)
	if err != nil {
		return err
	}
	err = RDB.Set(ctx, rdbKey, nil, 0).Err()
	log.Debug().Err(err).Msg("CACHE: generated person settings in db, caching")
	return err
}

// SettingPersonGet returns the value of col in table for the specified person
// in the specified place.
func (db *SQLDB) SettingPersonGet(col string, person, place int64) (any, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return false, err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	// Make sure that the person settings are present
	if err := SettingsPersonGenerate(tx, person, place); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM settings_person
		WHERE person = $1 and place = $2
	`, col)
	row := tx.QueryRow(query, person, place)

	var val any
	err = row.Scan(&val)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("got value")

	if err != nil {
		return nil, err
	}
	return val, tx.Commit()
}

// SettingPersonSet sets the value of col in table for the specified person in
// the specified place.
func (db *SQLDB) SettingPersonSet(col string, person, place int64, val any) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err := tx.Rollback(); err != nil {
			log.Debug().Err(err).Msg("failed to rollback transaction")
		}
	}(tx)

	// Make sure that the person settings are present
	if err := SettingsPersonGenerate(tx, person, place); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE settings_person
		SET %s = $1
		WHERE person = $2 and place = $3
	`, col)
	_, err = tx.Exec(query, val, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("changed setting")

	if err != nil {
		return err
	}
	return tx.Commit()
}
