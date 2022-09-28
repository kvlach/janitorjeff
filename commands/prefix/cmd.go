package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Command = &core.CommandStatic{
	Names: []string{
		"prefix",
	},
	Description: "Add, delete, list or reset prefixes.",
	UsageArgs:   "(add|del) <prefix> | list | reset",
	Target:      0, // TODO: change this to platforms.All
	Run:         run,

	Children: core.Commands{
		cmdAdd,
		cmdDel,
		cmdList,
		cmdReset,
	},
}

var cmdAdd = &core.CommandStatic{
	Names: []string{
		"add",
		"new",
	},
	Description: "Add a prefix.",
	UsageArgs:   "<prefix>",
	Run:         runAdd,
}

var cmdDel = &core.CommandStatic{
	Names: []string{
		"del",
		"delete",
		"rm",
		"remove",
	},
	Description: "Delete a prefix.",
	UsageArgs:   "<prefix>",
	Run:         runDelete,
}

var cmdList = &core.CommandStatic{
	Names: []string{
		"list",
		"ls",
	},
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
