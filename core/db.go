package core

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var DB *SQLDB

const schema = `
CREATE TABLE IF NOT EXISTS Scopes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	frontend INTEGER NOT NULL,
	original_id VARCHAR(255) NOT NULL
);
-- Info about why this is here in discord's Scope() implementation
INSERT OR IGNORE INTO Scopes VALUES(1,1,'');

CREATE TABLE IF NOT EXISTS CommandPrefixPrefixes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	place INTEGER NOT NULL,
	prefix VARCHAR(255) NOT NULL,
	type INTEGER NOT NULL,
	UNIQUE(place, prefix),
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
);
`

type SQLDB struct {
	Lock sync.RWMutex
	DB   *sql.DB
}

func Open(driver, source string) (*SQLDB, error) {
	sqlDB, err := sql.Open(driver, fmt.Sprintf("file:%s?_foreign_keys=on", source))
	if err != nil {
		return nil, err
	}

	db := &SQLDB{DB: sqlDB}
	if err := db.Init(schema); err != nil {
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

func (_ *SQLDB) ScopeAdd(tx *sql.Tx, id string, frontend int) (int64, error) {
	res, err := tx.Exec(`
		INSERT INTO Scopes(original_id, frontend)
		VALUES (?, ?)`, id, frontend)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// Returns the given scope's frontend specific ID
func (db *SQLDB) ScopeID(scope int64) (string, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	var id string
	row := db.DB.QueryRow(`
		SELECT original_id
		FROM Scopes
		WHERE id = ?
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
		SELECT frontend
		FROM Scopes
		WHERE id = ?
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
		FROM CommandPrefixPrefixes
		WHERE place = ?`, place)
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

	return prefixes, rows.Err()
}
