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
	name := fmt.Sprintf("'%s'", strings.Join(m.Command.Runtime.Name, " "))
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
	err := f.FlagSet.Parse(f.Msg.Command.Runtime.Args)
	if err != nil {
		err = ErrSilence
	}
	return f.FlagSet.Args(), err
}

func TypeFlag(p *int, value int, f *Flags) {
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

func ScopeFlag(p *int64, f *Flags) {
	scope, err := f.Msg.ScopePlace()
	if err != nil {
		// todo
	}
	usage := "provide a scope, default value is the current scope"
	f.FlagSet.Int64Var(p, "scope", scope, usage)
}
