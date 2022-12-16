package prefix

import (
	"errors"
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"
	"github.com/janitorjeff/jeff-bot/frontends/discord"
)

func getAdminFlags(m *core.Message) (*flags, []string, error) {
	f := newFlags(m).TypeFlag().ScopeFlag()
	args, err := f.fs.Parse()
	return f, args, err
}

var (
	errAdmin           = errors.New("prefixes of type 'admin' are static, change the bot's config instead")
	errMoreThanOneType = errors.New("only one type per prefix allowed")
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.Message) bool {
	return true
}

func (admin) Names() []string {
	return []string{
		"prefix",
	}
}

func (admin) Description() string {
	return ""
}

func (c admin) UsageArgs() string {
	return c.Children().Usage()
}

func (admin) Parent() core.CommandStatic {
	return nil
}

func (admin) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdminAdd,
		AdminDelete,
		AdminList,
		AdminReset,
	}
}

func (admin) Init() error {
	return nil
}

func (admin) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

/////////
//     //
// add //
//     //
/////////

var AdminAdd = adminAdd{}

type adminAdd struct{}

func (c adminAdd) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminAdd) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminAdd) Names() []string {
	return core.Add
}

func (adminAdd) Description() string {
	return "add prefix"
}

func (adminAdd) UsageArgs() string {
	return ""
}

func (adminAdd) Parent() core.CommandStatic {
	return Admin
}

func (adminAdd) Children() core.CommandsStatic {
	return nil
}

func (adminAdd) Init() error {
	return nil
}

func (c adminAdd) Run(m *core.Message) (any, error, error) {
	prefix, collision, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend {
	case frontends.Discord:
		prefix = discord.PlaceInBackticks(prefix)
		collision = discord.PlaceInBackticks(collision)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
		collision = fmt.Sprintf("'%s'", collision)
	}

	return c.err(usrErr, m, prefix, collision), usrErr, nil
}

func (adminAdd) err(usrErr error, m *core.Message, prefix, collision string) any {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case core.ErrMissingArgs:
		return m.Usage()
	case ErrExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case ErrCustomCommandExists:
		return fmt.Sprintf("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.", prefix, collision)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminAdd) core(m *core.Message) (string, string, error, error) {
	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", "", nil, err
	}

	t := fs.typeFlag
	if core.OnlyOneBitSet(int(t)) == false {
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

	collision, usrErr, err := Add(prefix, t, scope)
	return prefix, collision, usrErr, err
}

////////////
//        //
// delete //
//        //
////////////

var AdminDelete = adminDelete{}

type adminDelete struct{}

func (c adminDelete) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminDelete) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (c adminDelete) Names() []string {
	return core.Delete
}

func (adminDelete) Description() string {
	return "add prefix"
}

func (adminDelete) UsageArgs() string {
	return ""
}

func (adminDelete) Parent() core.CommandStatic {
	return Admin
}

func (adminDelete) Children() core.CommandsStatic {
	return nil
}

func (adminDelete) Init() error {
	return nil
}

func (c adminDelete) Run(m *core.Message) (any, error, error) {
	prefix, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend {
	case frontends.Discord:
		prefix = discord.PlaceInBackticks(prefix)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
	}

	return c.err(usrErr, m, prefix), usrErr, nil
}

func (adminDelete) err(usrErr error, m *core.Message, prefix string) any {
	switch usrErr {
	case nil:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case core.ErrMissingArgs:
		return m.Usage()
	case ErrNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case ErrOneLeft:
		resetCmd := core.Format(AdminReset, m.Command.Prefix)
		switch m.Frontend {
		case frontends.Discord:
			resetCmd = discord.PlaceInBackticks(resetCmd)
		default:
			resetCmd = fmt.Sprintf("'%s'", resetCmd)
		}
		return fmt.Sprintf("Can't delete, %s is the only prefix left. If you wish to reset to the default prefixes run: %s", prefix, resetCmd)
	default:
		return fmt.Sprint(usrErr)
	}
}

func (adminDelete) core(m *core.Message) (string, error, error) {
	if len(m.Command.Args) < 1 {
		return "", core.ErrMissingArgs, nil
	}

	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", nil, err
	}

	scope := fs.scopeFlag
	prefix := args[0]

	usrErr, err := Delete(prefix, core.All, scope)
	return prefix, usrErr, err
}

//////////
//      //
// list //
//      //
//////////

var AdminList = adminList{}

type adminList struct{}

func (c adminList) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminList) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminList) Names() []string {
	return core.List
}

func (adminList) Description() string {
	return "list prefixes"
}

func (adminList) UsageArgs() string {
	return ""
}

func (adminList) Parent() core.CommandStatic {
	return Admin
}

func (adminList) Children() core.CommandsStatic {
	return nil
}

func (adminList) Init() error {
	return nil
}

func (c adminList) Run(m *core.Message) (any, error, error) {
	prefixes, err := c.core(m)
	return fmt.Sprint(prefixes), nil, err
}

func (adminList) core(m *core.Message) ([]core.Prefix, error) {
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return nil, err
	}
	return List(fs.typeFlag, fs.scopeFlag)
}

///////////
//       //
// reset //
//       //
///////////

var AdminReset = adminReset{}

type adminReset struct{}

func (c adminReset) Type() core.CommandType {
	return c.Parent().Type()
}

func (c adminReset) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (adminReset) Names() []string {
	return []string{
		"reset",
	}
}

func (adminReset) Description() string {
	return "reset prefixes"
}

func (adminReset) UsageArgs() string {
	return ""
}

func (adminReset) Parent() core.CommandStatic {
	return Admin
}

func (adminReset) Children() core.CommandsStatic {
	return nil
}

func (adminReset) Init() error {
	return nil
}

func (c adminReset) Run(m *core.Message) (any, error, error) {
	return "Reset prefixes.", nil, c.core(m)
}

func (adminReset) core(m *core.Message) error {
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return err
	}
	return Reset(fs.scopeFlag)
}
