package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Command = &core.CommandStatic{
	Names: []string{
		"prefix",
	},
	Description: "Add, delete, list or reset prefixes.",
	UsageArgs:   "(add|del) <prefix> | list | reset",
	Frontends:   frontends.All,
	Run:         run,
	Init:        init_,

	Children: core.Commands{
		cmdAdd,
		cmdDel,
		cmdList,
		cmdReset,
	},
}

var cmdAdd = &core.CommandStatic{
	Names:       core.Add,
	Description: "Add a prefix.",
	UsageArgs:   "<prefix>",
	Run:         runAdd,
}

var cmdDel = &core.CommandStatic{
	Names:       core.Delete,
	Description: "Delete a prefix.",
	UsageArgs:   "<prefix>",
	Run:         runDelete,
}

var cmdList = &core.CommandStatic{
	Names:       core.List,
	Description: "List the current prefixes.",
	UsageArgs:   "",
	Run:         runList,
}

var cmdReset = &core.CommandStatic{
	Names: []string{
		"reset",
	},
	Description: "Reset prefixes to bot defaults.",
	UsageArgs:   "",
	Run:         runReset,
}
