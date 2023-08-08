package youtube

import (
	"git.sr.ht/~slowtyper/janitorjeff/core"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Permitted(m *core.Message) bool {
	return AdvancedSearchVideo.Permitted(m)
}

func (normal) Names() []string {
	return Advanced.Names()
}

func (normal) Description() string {
	return AdvancedSearchVideo.Description()
}

func (normal) UsageArgs() string {
	return AdvancedSearchVideo.UsageArgs()
}

func (normal) Category() core.CommandCategory {
	return Advanced.Category()
}

func (normal) Examples() []string {
	return AdvancedSearchVideo.Examples()
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return nil
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, core.Urr, error) {
	return AdvancedSearchVideo.Run(m)
}
