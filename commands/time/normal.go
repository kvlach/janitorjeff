package time

import (
	"regexp"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/rs/zerolog/log"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(*core.Message) bool {
	return true
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return "Time stuff and things."
}

func (normal) UsageArgs() string {
	return "[<user> | (zone)]"
}

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalZone,
	}
}

func (c normal) Init() error {
	core.Hooks.Register(c.addReminder)
	return nil
}

func (normal) addReminder(m *core.Message) {
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

func (normal) Run(m *core.Message) (any, error, error) {
	return AdvancedNow.Run(m)
}

//////////
//      //
// zone //
//      //
//////////

var NormalZone = normalZone{}

type normalZone struct{}

func (c normalZone) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalZone) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalZone) Names() []string {
	return []string{
		"zone",
	}
}

func (normalZone) Description() string {
	return "Set or view your own timezone."
}

func (normalZone) UsageArgs() string {
	return "[timezone]"
}

func (c normalZone) Category() core.CommandCategory {
	return c.Parent().Category()
}

func (normalZone) Parent() core.CommandStatic {
	return Normal
}

func (normalZone) Children() core.CommandsStatic {
	return nil
}

func (normalZone) Init() error {
	return nil
}

func (normalZone) Run(m *core.Message) (any, error, error) {
	if len(m.Command.Args) == 0 {
		return AdvancedTimezoneShow.Run(m)
	}
	return AdvancedTimezoneSet.Run(m)
}
