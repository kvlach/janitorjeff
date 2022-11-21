package scope

import (
	"github.com/janitorjeff/jeff-bot/core"
)

func Place(target string, client core.Messenger) (int64, error) {
	id, err := client.PlaceID(target)
	if err != nil {
		return -1, err
	}
	return client.PlaceLogical(id)
}

func Person(target, parent string, client core.Messenger) (int64, error) {
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
