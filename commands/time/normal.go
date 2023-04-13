package time

import (
	"regexp"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/rs/zerolog/log"
)

//////////
//      //
// time //
//      //
//////////

var NormalTime = normalTime{}

type normalTime struct{}

func (normalTime) Type() core.CommandType {
	return core.Normal
}

func (normalTime) Permitted(*core.Message) bool {
	return true
}

func (normalTime) Names() []string {
	return Advanced.Names()
}

func (normalTime) Description() string {
	return "Time stuff and things."
}

func (normalTime) UsageArgs() string {
	return "[user]"
}

func (normalTime) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normalTime) Examples() []string {
	return nil
}

func (normalTime) Parent() core.CommandStatic {
	return nil
}

func (normalTime) Children() core.CommandsStatic {
	return nil
}

func (c normalTime) Init() error {
	core.Hooks.Register(c.addReminder)
	return nil
}

func (normalTime) addReminder(m *core.Message) {
	re := regexp.MustCompile(`^remind\s+me\s+to\s+` + `(?P<cmd>.+(in|on)\s+.+)`)

	if !re.MatchString(m.Raw) {
		return
	}

	groupNames := re.SubexpNames()
	for _, match := range re.FindAllStringSubmatch(m.Raw, -1) {
		for i, text := range match {
			if groupNames[i] == "cmd" {
				m.Raw = text
			}
		}
	}

	m.Command = &core.Command{
		CommandRuntime: core.CommandRuntime{
			Args: m.Fields(),
		},
	}

	resp, usrErr, err := AdvancedRemindAdd.Run(m)
	if err != nil {
		log.Debug().Err(err).Msg("failed to create reminder")
		return
	}
	m.Write(resp, usrErr)
}

func (normalTime) Run(m *core.Message) (any, error, error) {
	return AdvancedNow.Run(m)
}

//////////////
//          //
// timezone //
//          //
//////////////

var NormalTimezone = normalTimezone{}

type normalTimezone struct{}

func (c normalTimezone) Type() core.CommandType {
	return core.Normal
}

func (c normalTimezone) Permitted(m *core.Message) bool {
	return AdvancedTimezone.Permitted(m)
}

func (normalTimezone) Names() []string {
	return []string{
		"timezone",
		"tz",
	}
}

func (normalTimezone) Description() string {
	return "Set or view your own timezone."
}

func (normalTimezone) UsageArgs() string {
	return "[timezone]"
}

func (c normalTimezone) Category() core.CommandCategory {
	return AdvancedTimezone.Category()
}

func (normalTimezone) Examples() []string {
	return nil
}

func (normalTimezone) Parent() core.CommandStatic {
	return nil
}

func (normalTimezone) Children() core.CommandsStatic {
	return nil
}

func (normalTimezone) Init() error {
	return nil
}

func (normalTimezone) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedTimezoneShow.Run(m)
	}
	return AdvancedTimezoneSet.Run(m)
}
