package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var DB *SQLDB

var ctx = context.Background()

var UrrValNil = UrrNew("Value doesn't exist.")

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
	//goland:noinspection GoUnhandledErrorResult
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

type Val struct {
	val any
	err error
}

func (v Val) Bool() (bool, error) {
	if v.err != nil {
		return false, v.err
	}
	if v.val == nil {
		return false, errors.New("expected bool, got nil")
	}
	return v.val.(bool), nil
}

func (v Val) Int() (int, error) {
	if v.err != nil {
		return 0, v.err
	}
	if v.val == nil {
		return 0, errors.New("expected int, got nil")
	}
	return int(v.val.(int64)), nil
}

func (v Val) Int64() (int64, error) {
	if v.err != nil {
		return 0, v.err
	}
	if v.val == nil {
		return 0, errors.New("expected int64, got nil")
	}
	return v.val.(int64), nil
}

func (v Val) Str() (string, error) {
	if v.err != nil {
		return "", v.err
	}
	if v.val == nil {
		return "", errors.New("expected string, got nil")
	}
	return v.val.(string), nil
}

func (v Val) StrNil() (string, Urr, error) {
	if v.err != nil {
		return "", nil, v.err
	}
	if v.val == nil {
		return "", UrrValNil, nil
	}
	return v.val.(string), nil, nil
}

// Time returns a time object with the timezone set to UTC.
func (v Val) Time() (time.Time, error) {
	if v.err != nil {
		return time.Time{}, v.err
	}
	if v.val == nil {
		return time.Time{}, errors.New("expected time.Time, got nil")
	}
	return time.Unix(v.val.(int64), 0).UTC(), nil
}

func (v Val) Duration() (time.Duration, error) {
	if v.err != nil {
		return 0, v.err
	}
	if v.val == nil {
		return 0, errors.New("expected time.Duration, got nil")
	}
	return time.Duration(v.val.(int64)) * time.Second, nil
}

func (v Val) UUIDNil() (uuid.UUID, Urr, error) {
	if v.err != nil {
		return uuid.UUID{}, nil, v.err
	}
	if v.val == nil {
		return uuid.UUID{}, UrrValNil, nil
	}
	u, err := uuid.Parse(string(v.val.([]uint8)))
	return u, UrrValNil, err
}

type coords struct {
	person int64
	place  int64
}

type Tx struct {
	Tx     *sql.Tx
	db     *SQLDB
	person map[coords]struct{}
	place  map[int64]struct{}
}

func (db *SQLDB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, db: db}, nil
}

func (tx *Tx) Commit() error {
	log.Debug().Msg("POSTGRES: committing transaction")
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback()
}

////////////////////
//                //
// place settings //
//                //
////////////////////

func (tx *Tx) placeExists(place int64) (bool, error) {
	var exists bool

	row := tx.Tx.QueryRow(`
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
		Msg("POSTGRES: checked if place settings exist")

	return exists, err
}

func (tx *Tx) placeGenerate(place int64) error {
	_, err := tx.Tx.Exec(`
		INSERT INTO settings_place (place)
		VALUES ($1)
	`, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Msg("POSTGRES: generated place settings")

	return err
}

// PlaceEnsure will check if settings for the specified place exist and if not
// will generate them.
func (tx *Tx) PlaceEnsure(place int64) error {
	slog := log.With().Int64("place", place).Logger()

	if tx.place == nil {
		tx.place = make(map[int64]struct{})
	}

	if _, ok := tx.place[place]; ok {
		slog.Debug().Msg("LOCAL: place settings already generated, skipping")
		return nil
	}

	rdbKey := fmt.Sprintf("settings_place_%d", place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		slog.Debug().Msg("REDIS: place settings already generated, skipping")
		slog.Debug().Msg("LOCAL: caching place settings")
		tx.place[place] = struct{}{}
		return nil
	}

	exists, err := tx.placeExists(place)
	if err != nil {
		return err
	}
	if exists {
		err := RDB.Set(ctx, rdbKey, nil, 0).Err()
		slog.Debug().
			Err(err).
			Msg("REDIS: caching place settings")
		slog.Debug().Msg("LOCAL: caching place settings")
		tx.place[place] = struct{}{}
		return err
	}

	err = tx.placeGenerate(place)
	if err != nil {
		return err
	}
	slog.Debug().Msg("LOCAL: caching newly generated place settings")
	tx.place[place] = struct{}{}

	err = RDB.Set(ctx, rdbKey, nil, 0).Err()
	log.Debug().
		Err(err).
		Msg("REDIS: caching newly generated place settings")
	return err
}

// PlaceGet returns the value of col in the table for the specified place.
func (tx *Tx) PlaceGet(col string, place int64) Val {
	// Make sure that the place settings are present
	if err := tx.PlaceEnsure(place); err != nil {
		return Val{}
	}

	query := fmt.Sprintf(`SELECT %s FROM settings_place WHERE place = $1`, col)

	var val any
	err := tx.Tx.QueryRow(query, place).Scan(&val)
	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("POSTGRES: got value")
	return Val{val, err}
}

// PlaceSet sets the value of col in the table for the specified place.
func (tx *Tx) PlaceSet(col string, place int64, val any) error {
	// Make sure that the place settings are present
	if err := tx.PlaceEnsure(place); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE settings_place
		SET %s = $1
		WHERE place = $2
	`, col)
	_, err := tx.Tx.Exec(query, val, place)

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface(col, val).
		Msg("POSTGRES: changed setting")

	return err
}

func (db *SQLDB) PlaceGet(col string, place int64) Val {
	tx, err := db.Begin()
	if err != nil {
		return Val{}
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	val := tx.PlaceGet(col, place)
	if val.err != nil {
		return Val{}
	}
	return Val{val.val, tx.Commit()}
}

func (db *SQLDB) PlaceSet(col string, place int64, val any) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	err = tx.PlaceSet(col, place, val)
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

func (tx *Tx) personExists(person, place int64) (bool, error) {
	var exists bool

	err := tx.Tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM settings_person
			WHERE person = $1 and place = $2
			LIMIT 1
		)
	`, person, place).Scan(&exists)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Bool("exist", exists).
		Msg("POSTGRES: checked if person settings exist")

	return exists, err
}

func (tx *Tx) personGenerate(person, place int64) error {
	_, err := tx.Tx.Exec(`
		INSERT INTO settings_person (person, place)
		VALUES ($1, $2)
	`, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Msg("POSTGRES: generated person settings")

	return err
}

// PersonEnsure will check if settings for the specified person in the
// specified place exist, and if not will generate them.
func (tx *Tx) PersonEnsure(person, place int64) error {
	slog := log.With().
		Int64("person", person).
		Int64("place", place).
		Logger()

	if tx.person == nil {
		tx.person = make(map[coords]struct{})
	}

	if _, ok := tx.person[coords{person: person, place: place}]; ok {
		slog.Debug().Msg("LOCAL: person settings already generated, skipping")
		return nil
	}

	rdbKey := fmt.Sprintf("settings_person_%d_%d", person, place)

	if _, err := RDB.Get(ctx, rdbKey).Result(); err == nil {
		slog.Debug().Msg("REDIS: person settings already generated, skipping")
		slog.Debug().Msg("LOCAL: caching person settings")
		tx.person[coords{person: person, place: place}] = struct{}{}
		return nil
	}

	exists, err := tx.personExists(person, place)
	if err != nil {
		return err
	}
	if exists {
		err := RDB.Set(ctx, rdbKey, nil, 0).Err()
		slog.Debug().
			Err(err).
			Msg("REDIS: caching person settings")
		slog.Debug().Msg("LOCAL: caching person settings")
		tx.person[coords{person: person, place: place}] = struct{}{}
		return err
	}

	err = tx.personGenerate(person, place)
	if err != nil {
		return err
	}
	slog.Debug().Msg("LOCAL: caching newly generated person settings")
	tx.person[coords{person: person, place: place}] = struct{}{}

	err = RDB.Set(ctx, rdbKey, nil, 0).Err()
	slog.Debug().
		Err(err).
		Msg("REDIS: caching newly generated person settings")
	return err
}

// PersonGet returns the value of col in the table for the specified person
// in the specified place.
func (tx *Tx) PersonGet(col string, person, place int64) Val {
	// Make sure that the person settings are present
	if err := tx.PersonEnsure(person, place); err != nil {
		return Val{}
	}

	query := fmt.Sprintf(`SELECT %s FROM settings_person WHERE person = $1 and place = $2`, col)
	row := tx.Tx.QueryRow(query, person, place)

	var val any
	err := row.Scan(&val)
	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("POSTGRES: got value")
	return Val{val, err}
}

// PersonSet sets the value of col in the table for the specified person in
// the specified place.
func (tx *Tx) PersonSet(col string, person, place int64, val any) error {
	// Make sure that the person settings are present
	if err := tx.PersonEnsure(person, place); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE settings_person
		SET %s = $1
		WHERE person = $2 and place = $3
	`, col)
	_, err := tx.Tx.Exec(query, val, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Interface(col, val).
		Msg("POSTGRES: changed setting")

	return err
}

func (db *SQLDB) PersonGet(col string, person, place int64) Val {
	tx, err := db.Begin()
	if err != nil {
		return Val{}
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	val := tx.PersonGet(col, person, place)
	if val.err != nil {
		return Val{}
	}
	return Val{val.val, tx.Commit()}
}

func (db *SQLDB) PersonSet(col string, person, place int64, val any) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	err = tx.PersonSet(col, person, place, val)
	if err != nil {
		return err
	}
	return tx.Commit()
}
