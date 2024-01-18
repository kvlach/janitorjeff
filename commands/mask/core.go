package mask

import (
	"github.com/kvlach/janitorjeff/core"

	"github.com/kvlach/gosafe"
)

var UrrMaskNotSet = core.UrrNew("No mask has been set.")

type Target struct {
	Person int64
	Place  int64
}

// Only a handful of people are ever expected to be bot admins, so using a map
// should cause no problems.
var masks = gosafe.Map[int64, Target]{}

func Show(person int64) (Target, error, error) {
	t, ok := masks.Get(person)
	if !ok {
		return Target{}, UrrMaskNotSet, nil
	}
	return t, nil, nil
}

func Set(f core.Frontender, m core.Messenger, person int64, userID string, locID string) (Target, error) {
	person, err := m.Person(userID)
	if err != nil {
		return Target{}, err
	}

	place, err := f.PlaceLogical(locID)
	if err != nil {
		return Target{}, err
	}

	t := Target{
		Person: person,
		Place:  place,
	}

	masks.Set(person, t)
	return t, nil

}

func Delete(person int64) core.Urr {
	if _, ok := masks.Get(person); !ok {
		return UrrMaskNotSet
	}
	masks.Delete(person)
	return nil
}
