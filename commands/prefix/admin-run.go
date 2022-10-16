package prefix

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"
)

var (
	errAdmin           = errors.New("prefixes of type 'admin' are static, change the bot's config instead")
	errMoreThanOneType = errors.New("only one type per prefix allowed")
)

type flags struct {
	fs        *flag.FlagSet
	m         *core.Message
	typeFlag  int
	scopeFlag int64

	send strings.Builder
}

func (f *flags) Write(p []byte) (int, error) {
	return f.send.Write(p)
}

func (f *flags) defaultUsage() {
	if f.fs.Name() == "" {
		f.send.WriteString("Usage:\n")
	} else {
		fmt.Fprintf(&f.send, "Usage of %s:\n", f.fs.Name())
	}
	f.fs.PrintDefaults()

	f.send.WriteString("```")

	f.m.Write(f.send.String(), errors.New("flags error"))
}

func newFlags(m *core.Message) *flags {
	name := fmt.Sprintf("'%s'", strings.Join(m.Command.Runtime.Name, " "))
	f := &flags{
		fs: flag.NewFlagSet(name, flag.ContinueOnError),
		m:  m,
	}

	f.send.WriteString("```\n")
	f.fs.SetOutput(f)

	f.fs.Usage = f.defaultUsage
	return f
}

func (f *flags) TypeFlag() *flags {
	f.typeFlag = core.All

	f.fs.Func("type", "comma separated command types", func(s string) error {
		split := strings.Split(s, ",")

		if len(split) != 0 {
			f.typeFlag = 0
		}

		for _, opt := range split {
			switch opt {
			case "normal":
				f.typeFlag |= core.Normal
			case "advanced":
				f.typeFlag |= core.Advanced
			case "admin":
				f.typeFlag |= core.Admin
			default:
				return fmt.Errorf("invalid type '%s'", opt)
			}
		}

		return nil
	})

	return f
}

func (f *flags) ScopeFlag() *flags {
	scope, err := f.m.ScopePlace()
	if err != nil {
		// todo
	}
	usage := "provide a scope, default value is the current scope"
	f.fs.Int64Var(&f.scopeFlag, "scope", scope, usage)
	return f
}

func (f *flags) Parse() (*flags, []string, error) {
	err := f.fs.Parse(f.m.Command.Runtime.Args)
	return f, f.fs.Args(), err
}

func getFlags(m *core.Message) (*flags, []string, error) {
	// f, args, err := newFlags(m).TypeFlag().ScopeFlag().Parse()
	// return f.typeFlag, args, err
	return newFlags(m).TypeFlag().ScopeFlag().Parse()
}

func runAdminAdd(m *core.Message) (any, error, error) {
	prefix, collision, usrErr, err := runAdminAddCore(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Type {
	case core.Discord:
		prefix = discord.PlaceInBackticks(prefix)
		collision = discord.PlaceInBackticks(collision)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
		collision = fmt.Sprintf("'%s'", collision)
	}

	return runAdminAddErr(usrErr, m, prefix, collision), usrErr, nil
}

func runAdminAddErr(usrErr error, m *core.Message, prefix, collision string) any {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case errMissingArgument:
		return m.ReplyUsage()
	case errExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case errCustomCommandExists:
		return fmt.Sprintf("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.", prefix, collision)
	default:
		return fmt.Sprint(usrErr)
	}
}

func onlyOneBitSet(n int) bool {
	// https://stackoverflow.com/a/28303898
	return n&(n-1) == 0
}

func runAdminAddCore(m *core.Message) (string, string, error, error) {
	fs, args, usrErr := getFlags(m)
	t := fs.typeFlag
	if usrErr != nil {
		return "", "", usrErr, nil
	}
	if onlyOneBitSet(t) == false {
		return "", "", errMoreThanOneType, nil
	}
	if t == core.Admin {
		return "", "", errAdmin, nil
	}

	if len(args) == 0 {
		return "", "", errMissingArgument, nil
	}
	prefix := args[0]

	scope := fs.scopeFlag
	// scope, err := m.Scope()
	// if err != nil {
	// 	return prefix, "", nil, err
	// }

	collision, err := customCommandCollision(m, prefix)
	if err != nil {
		return prefix, "", nil, err
	}
	if collision != "" {
		return prefix, collision, errCustomCommandExists, nil
	}

	prefixes, scopeExists, err := core.ScopePrefixes(scope)
	if err != nil {
		return prefix, "", nil, err
	}

	for _, p := range prefixes {
		if p.Prefix == prefix {
			return prefix, "", errExists, nil
		}
	}

	if scopeExists == false {
		for _, p := range prefixes {
			// admin prefixes are static
			if p.Type == core.Admin {
				continue
			}

			if err = dbAdd(p.Prefix, scope, p.Type); err != nil {
				return prefix, "", nil, err
			}
		}

	}

	return prefix, "", nil, dbAdd(prefix, scope, t)
}

func runAdminDel(m *core.Message) (any, error, error) {
	prefix, usrErr, err := runAdminDelCore(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Type {
	case core.Discord:
		prefix = discord.PlaceInBackticks(prefix)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
	}

	return runAdminDelErr(usrErr, m, prefix), usrErr, nil

}

func runAdminDelErr(usrErr error, m *core.Message, prefix string) any {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case errMissingArgument:
		return m.ReplyUsage()
	case errNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case errOneLeft:
		resetCmd := cmdAdminReset.Format(m.Command.Runtime.Prefix)
		switch m.Type {
		case core.Discord:
			resetCmd = discord.PlaceInBackticks(resetCmd)
		default:
			resetCmd = fmt.Sprintf("'%s'", resetCmd)
		}
		return fmt.Sprintf("Can't delete, %s is the only prefix left. If you wish to reset to the default prefixes run: %s", prefix, resetCmd)
	default:
		return fmt.Sprint(usrErr)
	}
}

func runAdminDelCore(m *core.Message) (string, error, error) {
	if len(m.Command.Runtime.Args) < 1 {
		return "", errMissingArgument, nil
	}

	fs, args, usrErr := getFlags(m)
	if usrErr != nil {
		return "", usrErr, nil
	}

	scope := fs.scopeFlag
	// scope, err := m.Scope()
	// if err != nil {
	// 	return prefix, nil, err
	// }

	prefix := args[0]

	// log.Debug().
	// 	Str("prefix", prefix).
	// 	Int64("scope", scope).
	// 	Send()

	prefixes, scopeExists, err := core.ScopePrefixes(scope)
	if err != nil {
		return prefix, nil, err
	}

	exists := false
	for _, p := range prefixes {
		if p.Prefix == prefix {
			// Admin commands are static, so if the user tries to delete an
			// admin command, throw an error.
			if p.Type == core.Admin {
				return prefix, errAdmin, nil
			}
			exists = true
		}

	}

	if exists == false {
		return prefix, errNotFound, nil
	}
	if len(prefixes) == 1 {
		return prefix, errOneLeft, nil
	}

	// If the scope doesn't exist then the default prefixes are being used and
	// they are not present in the DB. So if the user tries to delete one
	// nothing will happen. So we first add them all to the DB. Admin prefixes
	// are not scope specific and thus don't need to be added.
	if scopeExists == false {
		for _, p := range prefixes {
			if p.Type == core.Admin {
				continue
			}

			if err = dbAdd(p.Prefix, scope, p.Type); err != nil {
				return prefix, nil, err
			}
		}
	}

	return prefix, nil, dbDel(prefix, scope)

}

func runAdminList(m *core.Message) (any, error, error) {
	prefixes, usrErr, err := runAdminListCore(m)
	if err != nil {
		return nil, nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	return fmt.Sprint(prefixes), nil, nil
}

func runAdminListCore(m *core.Message) ([]core.Prefix, error, error) {
	fs, _, usrErr := getFlags(m)
	types := fs.typeFlag
	if usrErr != nil {
		return nil, usrErr, nil
	}

	scope := fs.scopeFlag

	prefixes, _, err := core.ScopePrefixes(scope)
	if err != nil {
		return nil, nil, err
	}

	requestedPrefixes := []core.Prefix{}
	for _, p := range prefixes {
		if types&p.Type != 0 {
			requestedPrefixes = append(requestedPrefixes, p)
		}
	}

	return requestedPrefixes, nil, nil
}

func runAdminReset(m *core.Message) (any, error, error) {
	usrErr, err := runAdminResetCore(m)
	if err != nil {
		return nil, nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	return "Reset prefixes", nil, nil
}

func runAdminResetCore(m *core.Message) (error, error) {
	fs, _, usrErr := getFlags(m)
	if usrErr != nil {
		return usrErr, nil
	}
	scope := fs.scopeFlag
	return nil, dbReset(scope)
}

func runAdmin(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), errMissingArgument, nil
}
