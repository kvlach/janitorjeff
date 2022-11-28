package youtube

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
)

var Advanced = advanced{}

type advanced struct{}

func (advanced) Type() core.CommandType {
	return core.Advanced
}

func (advanced) Permitted(*core.Message) bool {
	return true
}

func (advanced) Names() []string {
	return []string{
		"youtube",
		"yt",
	}
}

func (advanced) Description() string {
	return "Youtube related commands."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Parent() core.CommandStatic {
	return nil
}

func (advanced) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedSearch,
	}
}

func (advanced) Init() error {
	return nil
}

func (advanced) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////
//        //
// search //
//        //
////////////

var AdvancedSearch = advancedSearch{}

type advancedSearch struct{}

func (c advancedSearch) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSearch) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSearch) Names() []string {
	return core.Search
}

func (advancedSearch) Description() string {
	return "Search for a video."
}

func (advancedSearch) UsageArgs() string {
	return "<title>"
}

func (advancedSearch) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSearch) Children() core.CommandsStatic {
	return nil
}

func (advancedSearch) Init() error {
	return nil
}

func (c advancedSearch) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	vid, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}
	return c.err(usrErr, vid), usrErr, nil
}

func (advancedSearch) err(usrErr error, v Video) string {
	switch usrErr {
	case nil:
		return "https://youtu.be/" + v.id
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedSearch) core(m *core.Message) (Video, error, error) {
	return SearchVideo(m.RawArgs(0))
}
