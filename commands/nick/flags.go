package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

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

	// var err error
	// var author int64

	// author, err = f.fs.Msg.ScopeAuthor()
	// f.person = author

	// f.fs.FlagSet.Func("person", "person reference", func(s string) error {
	// 	if err != nil {
	// 		return err
	// 	}

	// 	var person int64

	// 	// if the -place flagged is passed then use its value instead of the
	// 	// current place value
	// 	if f.place != 0 {
	// 		person, err = ParseUser(f.fs.Msg, f.place, s)
	// 	} else {
	// 		person, err = ParseUserHere(f.fs.Msg, s)
	// 	}

	// 	if err != nil {
	// 		return fmt.Errorf("could not find user '%s'", s)
	// 	}
	// 	f.person = person
	// 	return nil
	// })

	// return f
}

func (f *flags) Place() *flags {
	core.FlagPlace(&f.place, f.fs)
	return f

	// 	var err error
	// 	var here int64

	// 	here, err = f.fs.Msg.ScopeHere()
	// 	f.place = here

	// 	f.fs.FlagSet.Func("place", "place reference", func(s string) error {
	// 		if err != nil {
	// 			return err
	// 		}

	// 		id, err := f.fs.Msg.Client.PlaceID(s)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		place, err := f.fs.Msg.Client.PlaceScope(id)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		f.place = place
	// 		return nil
	// 	})

	// 	return f
}
