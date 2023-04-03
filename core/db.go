package core

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
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

	schema, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}

	db := &SQLDB{DB: sqlDB}
	if err := db.Init(string(schema)); err != nil {
		return nil, err
	}

	return db, nil
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

// Returns the given scope's frontend specific ID
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

// Returns then given scope's frontend id
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

// Returns the list of all prefixes for a specific scope.
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

func (db *SQLDB) settingsPlaceExist(place int64) (bool, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var exists bool

	row := db.DB.QueryRow(`
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

func (db *SQLDB) settingsPlaceGenerate(place int64) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
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
func (db *SQLDB) SettingsPlaceGenerate(place int64) error {
	rdbKey := fmt.Sprintf("settings_place_%d", place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		log.Debug().Msg("CACHE: place settings already generated")
		return nil
	}

	exists, err := db.settingsPlaceExist(place)
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

	err = db.settingsPlaceGenerate(place)
	if err != nil {
		return err
	}
	err = RDB.Set(ctx, rdbKey, nil, 0).Err()
	log.Debug().Err(err).Msg("CACHE: generated place settings in db, caching")
	return err
}

// SettingPlaceGet returns the value of col in table for the specified place.
func (db *SQLDB) SettingPlaceGet(col string, place int64) (any, error) {
	// Make sure that the place settings are present
	if err := db.SettingsPlaceGenerate(place); err != nil {
		return nil, err
	}

	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var val any

	query := fmt.Sprintf(`
		SELECT %s
		FROM settings_place
		WHERE place = $1
	`, col)

	row := db.DB.QueryRow(query, place)

	err := row.Scan(&val)

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("got value")

	return val, err
}

// SettingPlaceSet sets the value of col in table for the specified place.
func (db *SQLDB) SettingPlaceSet(col string, place int64, val any) error {
	// Make sure that the place settings are present
	if err := db.SettingsPlaceGenerate(place); err != nil {
		return err
	}

	db.Lock.Lock()
	defer db.Lock.Unlock()

	query := fmt.Sprintf(`
		UPDATE settings_place
		SET %s = $1
		WHERE place = $2
	`, col)

	_, err := db.DB.Exec(query, val, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("changed setting")

	return err
}

/////////////////////
//                 //
// person settings //
//                 //
/////////////////////

func (db *SQLDB) settingsPersonExist(person, place int64) (bool, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var exists bool

	row := db.DB.QueryRow(`
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

func (db *SQLDB) settingsPersonGenerate(person, place int64) error {
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
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
func (db *SQLDB) SettingsPersonGenerate(person, place int64) error {
	rdbKey := fmt.Sprintf("settings_person_%d_%d", person, place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		log.Debug().Msg("CACHE: person settings already generated")
		return nil
	}

	exists, err := db.settingsPersonExist(person, place)
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

	err = db.settingsPersonGenerate(person, place)
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
	// Make sure that the person settings are present
	if err := db.SettingsPersonGenerate(person, place); err != nil {
		return nil, err
	}

	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var val any

	query := fmt.Sprintf(`
		SELECT %s
		FROM settings_person
		WHERE person = $1 and place = $2
	`, col)

	row := db.DB.QueryRow(query, person, place)

	err := row.Scan(&val)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("got value")

	return val, err
}

// PlaceSettingSet sets the value of col in table for the specified person in
// the specified place.
func (db *SQLDB) SettingPersonSet(col string, person, place int64, val any) error {
	// Make sure that the person settings are present
	if err := db.SettingsPersonGenerate(person, place); err != nil {
		return err
	}

	db.Lock.Lock()
	defer db.Lock.Unlock()

	query := fmt.Sprintf(`
		UPDATE settings_person
		SET %s = $1
		WHERE person = $2 and place = $3
	`, col)

	_, err := db.DB.Exec(query, val, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("changed setting")

	return err
}
