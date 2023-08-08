package youtube

import (
	"fmt"

	"git.sr.ht/~slowtyper/janitorjeff/apis/youtube"
	"git.sr.ht/~slowtyper/janitorjeff/core"
	"git.sr.ht/~slowtyper/janitorjeff/frontends/discord"

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
	return "YouTube related commands."
}

func (c advanced) UsageArgs() string {
	return c.Children().Usage()
}

func (advanced) Category() core.CommandCategory {
	return core.CommandCategoryServices
}

func (advanced) Examples() []string {
	return nil
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

func (advanced) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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
	return core.AliasesSearch
}

func (advancedSearch) Description() string {
	return "Group of various search related commands."
}

func (c advancedSearch) UsageArgs() string {
	return c.Children().Usage()
}

func (c advancedSearch) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSearch) Examples() []string {
	return nil
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

func (advancedSearch) Run(m *core.Message) (any, core.Urr, error) {
	return m.Usage(), core.UrrMissingArgs, nil
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
	return "Search for a YouTube video."
}

func (advancedSearchVideo) UsageArgs() string {
	return "<title>"
}

func (c advancedSearchVideo) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSearchVideo) Examples() []string {
	return []string{
		"gangnam style",
	}
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

func (c advancedSearchVideo) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSearchVideo) discord(m *core.Message) (any, core.Urr, error) {
	vid, urr, err := c.core(m)
	if err != nil {
		return nil, nil, err
	}

	// let discord handle the url embed
	if urr == nil {
		return vid.URL(), nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: fmt.Sprint(urr),
	}
	return embed, urr, nil
}

func (c advancedSearchVideo) text(m *core.Message) (string, core.Urr, error) {
	vid, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}
	return fmt.Sprintf("%s | %s", vid.Title, vid.URL()), nil, nil
}

func (advancedSearchVideo) core(m *core.Message) (youtube.Video, core.Urr, error) {
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
	return "Search for a YouTube channel."
}

func (advancedSearchChannel) UsageArgs() string {
	return "<channel name>"
}

func (c advancedSearchChannel) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (advancedSearchChannel) Examples() []string {
	return []string{
		"ben eater",
	}
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

func (c advancedSearchChannel) Run(m *core.Message) (any, core.Urr, error) {
	if len(m.Command.Args) < 1 {
		return m.Usage(), core.UrrMissingArgs, nil
	}

	switch m.Frontend.Type() {
	case discord.Frontend.Type():
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c advancedSearchChannel) discord(m *core.Message) (any, core.Urr, error) {
	ch, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}

	// let discord handle the url embed
	if urr == nil {
		return ch.URL(), nil, nil
	}

	embed := &dg.MessageEmbed{
		Description: fmt.Sprint(urr),
	}
	return embed, urr, nil
}

func (c advancedSearchChannel) text(m *core.Message) (string, core.Urr, error) {
	ch, urr, err := c.core(m)
	if err != nil {
		return "", nil, err
	}
	if urr != nil {
		return fmt.Sprint(urr), urr, nil
	}
	return fmt.Sprintf("%s | %s", ch.Title, ch.URL()), nil, nil
}

func (advancedSearchChannel) core(m *core.Message) (youtube.Channel, core.Urr, error) {
	return SearchChannel(m.RawArgs(0))
}
