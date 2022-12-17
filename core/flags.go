package core

import (
	"flag"
	"fmt"
	"strings"
)

type Flags struct {
	FlagSet *flag.FlagSet
	Msg     *Message
	send    strings.Builder
}

func NewFlags(m *Message) *Flags {
	name := fmt.Sprintf("'%s'", strings.Join(m.Command.Path, " "))
	f := &Flags{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
		Msg:     m,
	}

	f.send.WriteString("```\n")
	f.FlagSet.SetOutput(f)
	f.FlagSet.Usage = f.Usage

	return f
}

// required for the flagSet SetOutput option
func (f *Flags) Write(p []byte) (int, error) {
	return f.send.Write(p)
}

func (f *Flags) Usage() {
	if f.FlagSet.Name() == "" {
		f.send.WriteString("Usage:\n")
	} else {
		fmt.Fprintf(&f.send, "Usage of %s:\n", f.FlagSet.Name())
	}
	f.FlagSet.PrintDefaults()
	f.send.WriteString("```")
	f.Msg.Write(f.send.String(), nil)
}

func (f *Flags) Parse() ([]string, error) {
	err := f.FlagSet.Parse(f.Msg.Command.Args)
	if err != nil {
		err = ErrSilence
	}
	return f.FlagSet.Args(), err
}

func TypeFlag(p *CommandType, value CommandType, f *Flags) {
	*p = value

	f.FlagSet.Func("type", "comma separated command types", func(s string) error {
		split := strings.Split(s, ",")

		if len(split) == 0 {
			return nil
		}

		// we set to zero because we no longer want the default value
		*p = 0

		for _, opt := range split {
			switch opt {
			case "normal":
				*p |= Normal
			case "advanced":
				*p |= Advanced
			case "admin":
				*p |= Admin
			default:
				return fmt.Errorf("invalid type '%s'", opt)
			}
		}

		return nil
	})
}

func FlagPlace(place *int64, f *Flags) {
	here, err := f.Msg.Here.ScopeLogical()
	if err != nil {
		// todo
	}
	usage := "provide a place, default value is 'here'"
	f.FlagSet.Int64Var(place, "place", here, usage)
}

func FlagPerson(person *int64, f *Flags) {
	author, err := f.Msg.Author.Scope()
	if err != nil {
		// todo
	}
	usage := "provide a person, default value is 'author'"
	f.FlagSet.Int64Var(person, "person", author, usage)
}
