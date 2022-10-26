package scope

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

func runPlace(target string, client core.Messenger) (int64, error) {
	id, err := client.PlaceID(target)
	if err != nil {
		return -1, err
	}
	return client.PlaceLogical(id)
}

func runPerson(target, parent string, client core.Messenger) (int64, error) {
	placeID, err := client.PlaceID(parent)
	if err != nil {
		return -1, err
	}

	id, err := client.PersonID(target, placeID)
	if err != nil {
		return -1, err
	}

	return client.Person(id)
}
