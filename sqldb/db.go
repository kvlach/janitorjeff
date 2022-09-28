package sqldb

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

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

func (db *DB) PrefixList(scope int64) ([]string, error) {
	db.Lock.RLock()
	defer db.Lock.RUnlock()

	// We order by the length of the prefix in order to avoid matching the
	// wrong prefix. For example, in DMs the empty string prefix is added. If
	// it is placed first in the list of prefixes then it always get matched.
	// So even if the user uses for example `!`, the command will be parsed as
	// having the empty prefix and will fail to match (since it will try to
	// match the whole thing, `!test` for example, instead of trimming the
	// prefix first). This also can happen if for example there exist the `!!`
	// and `!` prefixes. If the single `!` is first on the list and the user
	// uses `!!` then the same problem occurs.
	rows, err := db.DB.Query(`
		SELECT prefix
		FROM CommandPrefixPrefixes
		WHERE scope = ?
		ORDER BY length(prefix) DESC`, scope)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []string
	for rows.Next() {
		var prefix string
		if err := rows.Scan(&prefix); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, rows.Err()
}
