package lens

import "github.com/kvlach/janitorjeff/core"

func DirectorAdd(name string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		INSERT INTO cmd_lens_directors(name)
		VALUES ($1)`, name)

	return err
}

func DirectorDelete(name string) error {
	db := core.DB
	db.Lock.Lock()
	defer db.Lock.Unlock()

	_, err := db.DB.Exec(`
		DELETE FROM cmd_lens_directors
		WHERE name = $1`, name)

	return err
}
