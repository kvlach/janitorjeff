package prefix

import (
	"fmt"

	"github.com/kvlach/janitorjeff/core"
	"github.com/kvlach/janitorjeff/frontends/discord"
)

func getAdminFlags(m *core.EventMessage) (*flags, []string, error) {
	f := newFlags(m).TypeFlag().ScopeFlag()
	args, err := f.fs.Parse()
	return f, args, err
}

var (
	UrrAdmin           = core.UrrNew("prefixes of type 'admin' are static, change the bot's config instead")
	UrrMoreThanOneType = core.UrrNew("only one type per prefix allowed")
)

var Admin = admin{}

type admin struct{}

func (admin) Type() core.CommandType {
	return core.Admin
}

func (admin) Permitted(*core.EventMessage) bool {
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

func (admin) Category() core.CommandCategory {
	return Advanced.Category()
}

func (admin) Examples() []string {
	return nil
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

func (admin) Run(m *core.EventMessage) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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

func (c adminAdd) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminAdd) Names() []string {
	return core.AliasesAdd
}

func (adminAdd) Description() string {
	return "add prefix"
}

func (adminAdd) UsageArgs() string {
	return ""
}

func (c adminAdd) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminAdd) Examples() []string {
	return nil
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

func (c adminAdd) Run(m *core.EventMessage) (any, core.Urr, error) {
	prefix, collision, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		prefix = discord.PlaceInBackticks(prefix)
		collision = discord.PlaceInBackticks(collision)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
		collision = fmt.Sprintf("'%s'", collision)
	}

	return c.fmt(urr, m, prefix, collision), urr, nil
}

func (adminAdd) fmt(urr core.Urr, m *core.EventMessage, prefix, collision string) any {
	switch urr {
	case nil:
		return fmt.Sprintf("Added prefix %s", prefix)
	case core.UrrMissingArgs:
		return m.Usage()
	case UrrExists:
		return fmt.Sprintf("Prefix %s already exists.", prefix)
	case UrrCustomCommandExists:
		return fmt.Sprintf("Can't add the prefix %s. A custom command with the name %s exists and would collide with the built-in command of the same name. Either change the custom command or use a different prefix.", prefix, collision)
	default:
		return fmt.Sprint(urr)
	}
}

func (adminAdd) core(m *core.EventMessage) (string, string, core.Urr, error) {
	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", "", nil, err
	}

	t := fs.typeFlag
	if core.OnlyOneBitSet(int(t)) == false {
		return "", "", UrrMoreThanOneType, nil
	}
	if t == core.Admin {
		return "", "", UrrAdmin, nil
	}

	if len(args) == 0 {
		return "", "", core.UrrMissingArgs, nil
	}
	prefix := args[0]

	scope := fs.scopeFlag

	collision, urr, err := Add(prefix, t, scope)
	return prefix, collision, urr, err
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

func (c adminDelete) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (c adminDelete) Names() []string {
	return core.AliasesDelete
}

func (adminDelete) Description() string {
	return "add prefix"
}

func (adminDelete) UsageArgs() string {
	return ""
}

func (c adminDelete) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminDelete) Examples() []string {
	return nil
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

func (c adminDelete) Run(m *core.EventMessage) (any, core.Urr, error) {
	prefix, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		prefix = discord.PlaceInBackticks(prefix)
	default:
		prefix = fmt.Sprintf("'%s'", prefix)
	}

	return c.fmt(urr, m, prefix), urr, nil
}

func (adminDelete) fmt(urr error, m *core.EventMessage, prefix string) any {
	switch urr {
	case nil:
		return fmt.Sprintf("Deleted prefix %s", prefix)
	case core.UrrMissingArgs:
		return m.Usage()
	case UrrNotFound:
		return fmt.Sprintf("Prefix %s doesn't exist.", prefix)
	case UrrOneLeft:
		resetCmd := core.Format(AdminReset, m.Command.Prefix)
		switch m.Frontend.Type() {
		case discord.Frontend.Type():
			resetCmd = discord.PlaceInBackticks(resetCmd)
		default:
			resetCmd = fmt.Sprintf("'%s'", resetCmd)
		}
		return fmt.Sprintf("Can't delete, %s is the only prefix left. If you wish to reset to the default prefixes run: %s", prefix, resetCmd)
	default:
		return fmt.Sprint(urr)
	}
}

func (adminDelete) core(m *core.EventMessage) (string, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return "", core.UrrMissingArgs, nil
	}

	fs, args, err := getAdminFlags(m)
	if err != nil {
		return "", nil, err
	}

	scope := fs.scopeFlag
	prefix := args[0]

	urr, err := Delete(prefix, core.All, scope)
	return prefix, urr, err
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

func (c adminList) Permitted(m *core.EventMessage) bool {
	return c.Parent().Permitted(m)
}

func (adminList) Names() []string {
	return core.AliasesList
}

func (adminList) Description() string {
	return "list prefixes"
}

func (adminList) UsageArgs() string {
	return ""
}

func (c adminList) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminList) Examples() []string {
	return nil
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

func (c adminList) Run(m *core.EventMessage) (any, core.Urr, error) {
	prefixes, err := c.core(m)
	return fmt.Sprint(prefixes), nil, err
}

func (adminList) core(m *core.EventMessage) ([]core.Prefix, error) {
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

func (c adminReset) Permitted(m *core.EventMessage) bool {
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

func (c adminReset) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (adminReset) Examples() []string {
	return nil
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

func (c adminReset) Run(m *core.EventMessage) (any, core.Urr, error) {
	return "Reset prefixes.", nil, c.core(m)
}

func (adminReset) core(m *core.EventMessage) error {
	fs, _, err := getAdminFlags(m)
	if err != nil {
		return err
	}
	return Reset(fs.scopeFlag)
}
