package youtube

import (
	"fmt"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends"

	dg "github.com/bwmarrin/discordgo"
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
	return "Group of various search realted commands."
}

func (c advancedSearch) UsageArgs() string {
	return c.Children().Usage()
}

func (advancedSearch) Parent() core.CommandStatic {
	return Advanced
}

func (advancedSearch) Children() core.CommandsStatic {
	return core.CommandsStatic{
		AdvancedSearchVideo,
		AdvancedSearchChannel,
	}
}

func (advancedSearch) Init() error {
	return nil
}

func (advancedSearch) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

//////////////////
//              //
// search video //
//              //
//////////////////

var AdvancedSearchVideo = advancedSearchVideo{}

type advancedSearchVideo struct{}

func (c advancedSearchVideo) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSearchVideo) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (advancedSearchVideo) Names() []string {
	return []string{
		"video",
		"vid",
	}
}

func (advancedSearchVideo) Description() string {
	return "Search for a video."
}

func (advancedSearchVideo) UsageArgs() string {
	return "<title>"
}

func (advancedSearchVideo) Parent() core.CommandStatic {
	return AdvancedSearch
}

func (advancedSearchVideo) Children() core.CommandsStatic {
	return nil
}

func (advancedSearchVideo) Init() error {
	return nil
}

func (c advancedSearchVideo) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSearchVideo) discord(m *core.Message) (any, error, error) {
	vid, usrErr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	// let discord handle the url embed
	if usrErr == nil {
		return vid.URL(), nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: fmt.Sprint(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedSearchVideo) text(m *core.Message) (string, error, error) {
	vid, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	return fmt.Sprintf("%s | %s", vid.Title, vid.URL()), nil, nil
}

func (advancedSearchVideo) core(m *core.Message) (Video, error, error) {
	return SearchVideo(m.RawArgs(0))
}

////////////////////
//                //
// search channel //
//                //
////////////////////

var AdvancedSearchChannel = advancedSearchChannel{}

type advancedSearchChannel struct{}

func (c advancedSearchChannel) Type() core.CommandType {
	return c.Parent().Type()
}

func (c advancedSearchChannel) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (c advancedSearchChannel) Names() []string {
	return []string{
		"channel",
		"ch",
	}
}

func (advancedSearchChannel) Description() string {
	return "Search for a channel."
}

func (advancedSearchChannel) UsageArgs() string {
	return "<channel...>"
}

func (advancedSearchChannel) Parent() core.CommandStatic {
	return AdvancedSearch
}

func (advancedSearchChannel) Children() core.CommandsStatic {
	return nil
}

func (advancedSearchChannel) Init() error {
	return nil
}

func (c advancedSearchChannel) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.ErrMissingArgs, nil
	}

	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSearchChannel) discord(m *core.Message) (any, error, error) {
	ch, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	// let disocrd handle the url embed
	if usrErr == nil {
		return ch.URL(), nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: fmt.Sprint(usrErr),
	}
	return embed, usrErr, nil
}

func (c advancedSearchChannel) text(m *core.Message) (string, error, error) {
	ch, usrErr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if usrErr != nil {
		return fmt.Sprint(usrErr), usrErr, nil
	}
	return fmt.Sprintf("%s | %s", ch.Title, ch.URL()), nil, nil
}

func (advancedSearchChannel) err(usrErr error, ch Channel) string {
	switch usrErr {
	case nil:
		return ch.URL()
	default:
		return fmt.Sprint(usrErr)
	}
}

func (advancedSearchChannel) core(m *core.Message) (Channel, error, error) {
	return SearchChannel(m.RawArgs(0))
}
