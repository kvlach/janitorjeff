package nick

import (
	"errors"

	"git.slowtyper.com/slowtyper/janitorjeff/core"

	"github.com/rs/zerolog/log"
)

var (
	errPersonNotFound = errors.New("user nick not found")
	errNickExists     = errors.New("nick is used by a different user")
)

//////////////
//          //
// database //
//          //
//////////////

const dbSchema = `
CREATE TABLE IF NOT EXISTS CommandNickNicknames (
	person INTEGER NOT NULL,
	place INTEGER NOT NULL,
	nick VARCHAR(255) NOT NULL,
	UNIQUE(person, place),
	UNIQUE(place, nick),
	FOREIGN KEY (person) REFERENCES Scopes(id) ON DELETE CASCADE,
	FOREIGN KEY (place) REFERENCES Scopes(id) ON DELETE CASCADE
)
`

func dbPersonAdd(person, place int64, nick string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
	INSERT INTO CommandNickNicknames(person, place, nick)
	VALUES (?, ?, ?)`, person, place, nick)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("nick", nick).
		Msg("added person nick in db")

	return err
}

func dbPersonUpdate(person, place int64, nick string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		UPDATE CommandNickNicknames
		SET nick = ?
		WHERE person = ? and place = ?
	`, nick, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("nick", nick).
		Msg("updated person nick")

	return err
}

func dbPersonExists(person, place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandNickNicknames
			WHERE person = ? and place = ?
			LIMIT 1
		)`, person, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if scope exists")

	return exists, err
}

func dbNickExists(nick string, place int64) (bool, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var exists bool

	row := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM CommandNickNicknames
			WHERE nick = ? and place = ?
			LIMIT 1
		)`, nick, place)

	err := row.Scan(&exists)

	log.Debug().
		Err(err).
		Str("nick", nick).
		Int64("place", place).
		Bool("exists", exists).
		Msg("checked db to see if nick exists")

	return exists, err
}

func dbPersonDelete(person, place int64) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM CommandNickNicknames
		WHERE person = ? and place = ?`, person, place)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Msg("deleted person from db")

	return err
}

func dbPersonNick(person, place int64) (string, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var nick string

	row := db.DB.QueryRow(`
		SELECT nick
		FROM CommandNickNicknames
		WHERE person = ? and place = ?`, person, place)

	err := row.Scan(&nick)

	log.Debug().
		Err(err).
		Int64("person", person).
		Int64("place", place).
		Str("nick", nick).
		Msg("got nick for person")

	return nick, err
}

func dbGetPerson(nick string, place int64) (int64, error) {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	var person int64

	row := db.DB.QueryRow(`
		SELECT person
		FROM CommandNickNicknames
		WHERE nick = ? and place = ?`, nick, place)

	err := row.Scan(&person)

	log.Debug().
		Err(err).
		Str("nick", nick).
		Int64("place", place).
		Int64("person", person).
		Msg("got nick for person")

	return person, err
}

/////////
//     //
// run //
//     //
/////////

func Show(person, place int64) (string, error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return "", errPersonNotFound, nil
	}

	nick, err := dbPersonNick(person, place)
	return nick, nil, err
}

func Set(nick string, person, place int64) (error, error) {
	nickExists, err := dbNickExists(nick, place)
	if err != nil {
		return nil, err
	}
	if nickExists {
		return errNickExists, nil
	}

	personExists, err := dbPersonExists(person, place)
	if err != nil {
		return nil, err
	}

	if personExists {
		return nil, dbPersonUpdate(person, place, nick)
	}
	return nil, dbPersonAdd(person, place, nick)
}

func Delete(person, place int64) (error, error) {
	exists, err := dbPersonExists(person, place)
	if err != nil {
		return nil, err
	}
	if !exists {
		return errPersonNotFound, nil
	}
	return nil, dbPersonDelete(person, place)
}

///////////
//       //
// flags //
//       //
///////////

type flags struct {
	fs *core.Flags

	person int64
	place  int64
}

func newFlags(m *core.Message) *flags {
	f := &flags{
		fs: core.NewFlags(m),
	}
	return f
}

func (f *flags) Person() *flags {
	core.FlagPerson(&f.person, f.fs)
	return f
}

func (f *flags) Place() *flags {
	core.FlagPlace(&f.place, f.fs)
	return f
}

//////////
//      //
// init //
//      //
//////////

func init_() error {
	return core.DB.Init(dbSchema)
}
