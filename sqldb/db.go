package sqldb

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Prefixes:
//
// Some very basic prefix support has to exist in the core in order to be able
// to find scope specific prefixes when parsing the message. Only thing that is
// supported is creating the table and getting a list of prefixes for the
// current scope. The rest (adding, deleting, etc.) is handled externally by a
// command.
//
// The idea of having the ability to have an arbitrary number of prefixes
// originated from the fact that I wanted to not have to use a prefix when
// DMing the bot, as it is pointless to do so in that case. So after building
// support for more than 1 prefix it felt unecessary to limit it to just DMs,
// as that would be an artificial limit on what the bot is already supports and
// not really a necessity.
//
// There are 3 main prefix types:
//  - normal
//  - advanced
//  - admin
// These are used to call normal, advanced and admin commands respectively. A
// prefix has to be unique accross all types (in a specific scope), for example
// the prefix `!` can't be used for both normal and advanced commands.

type Prefix struct {
	Type   int
	Prefix string
}

const schema = `
CREATE TABLE IF NOT EXISTS Scopes (
	id INTEGER PRIMARY KEY AUTOINCREMENT
);
-- Info about why this is here in discord's Scope() implementation
INSERT OR IGNORE INTO Scopes VALUES(1);

CREATE TABLE IF NOT EXISTS PlatformDiscordGuilds (
	id INTEGER PRIMARY KEY,
	guild VARCHAR(255) NOT NULL UNIQUE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);
-- Info about why this is here in discord's Scope() implementation
INSERT OR IGNORE INTO PlatformDiscordGuilds VALUES(1,'');

CREATE TABLE IF NOT EXISTS PlatformDiscordChannels (
	id INTEGER PRIMARY KEY,
	channel VARCHAR(255) NOT NULL UNIQUE,
	guild INTEGER NOT NULL,
	FOREIGN KEY (guild) REFERENCES PlatformDiscordGuilds(id) ON DELETE CASCADE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS PlatformDiscordUsers (
	id INTEGER PRIMARY KEY,
	user VARCHAR(255) NOT NULL UNIQUE,
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS PlatformTwitchChannels (
	id INTEGER PRIMARY KEY,
	channel_id VARCHAR(255) NOT NULL UNIQUE,
	channel_name VARCHAR(255) NOT NULL,
	access_token VARCHAR(255),
	refresh_token VARCHAR(255),
	FOREIGN KEY (id) REFERENCES Scopes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS CommandPrefixPrefixes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	scope INTEGER NOT NULL,
	prefix VARCHAR(255) NOT NULL,
	type INTEGER NOT NULL,
	UNIQUE(scope, prefix),
	FOREIGN KEY (scope) REFERENCES Scopes(id) ON DELETE CASCADE
);
`

type DB struct {
	Lock sync.RWMutex
	DB   *sql.DB
}

func Open(driver, source string) (*DB, error) {
	sqlDB, err := sql.Open(driver, fmt.Sprintf("file:%s?_foreign_keys=on", source))
	if err != nil {
		return nil, err
	}

	db := &DB{DB: sqlDB}
	if err := db.Init(schema); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	db.Lock.Lock()
	defer db.Lock.Unlock()
	return db.DB.Close()
}

func (db *DB) Init(schema string) error {
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

func (_ *DB) ScopeAdd(tx *sql.Tx) (int64, error) {
	// Must pass *a* value to create a new row, this will auto increment the id
	// as expected
	res, err := tx.Exec("INSERT INTO Scopes(id) VALUES (NULL)")
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// Returns the list of all prefixes for a specific scope.
func (db *DB) PrefixList(scope int64) ([]Prefix, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	rows, err := db.DB.Query(`
		SELECT prefix, type
		FROM CommandPrefixPrefixes
		WHERE scope = ?`, scope)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []Prefix
	for rows.Next() {
		var prefix string
		var type_ int
		if err := rows.Scan(&prefix, &type_); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, Prefix{Type: type_, Prefix: prefix})
	}

	return prefixes, rows.Err()
}
