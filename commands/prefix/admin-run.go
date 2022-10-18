package prefix

import (
	"errors"
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/discord"
)

var (
	errAdmin           = errors.New("prefixes of type 'admin' are static, change the bot's config instead")
	errMoreThanOneType = errors.New("only one type per prefix allowed")
)

type adminFlags struct {
	fs *core.Flags

	typeFlag  int
	scopeFlag int64
}

func (f *adminFlags) TypeFlag() *adminFlags {
	core.TypeFlag(&f.typeFlag, core.All, f.fs)
	return f
}

func (f *adminFlags) ScopeFlag() *adminFlags {
	core.ScopeFlag(&f.scopeFlag, f.fs)
	return f
}

func getAdminFlags(m *core.Message) (*adminFlags, []string, error) {
	f := &adminFlags{
		fs: core.NewFlags(m),
	}
	f.TypeFlag().ScopeFlag()
	args, err := f.fs.Parse()
	return f, args, err
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
	case core.ErrMissingArgs:
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
	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", "", nil, err
	}

	t := fs.typeFlag
	if onlyOneBitSet(t) == false {
		return "", "", errMoreThanOneType, nil
	}
	if t == core.Admin {
		return "", "", errAdmin, nil
	}

	if len(args) == 0 {
		return "", "", core.ErrMissingArgs, nil
	}
	prefix := args[0]

	scope := fs.scopeFlag

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
	case core.ErrMissingArgs:
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
		return "", core.ErrMissingArgs, nil
	}

	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", nil, err
	}

	scope := fs.scopeFlag
	prefix := args[0]

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
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return nil, nil, err
	}

	types := fs.typeFlag
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
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return nil, err
	}
	scope := fs.scopeFlag
	return nil, dbReset(scope)
}

func runAdmin(m *core.Message) (any, error, error) {
	return m.ReplyUsage(), core.ErrMissingArgs, nil
}
