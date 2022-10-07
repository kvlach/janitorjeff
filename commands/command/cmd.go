package command

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Command = &core.CommandStatic{
	Names: []string{
		"cmd",
		"command",
	},
	Description: "Add, delete, modify or list custom commands.",
	UsageArgs:   "(add|modify|del|list)",
	Run:         run,
	Init:        init_,

	Children: core.Commands{
		cmdAdd,
		cmdDel,
		cmdModify,
		cmdList,
		cmdHistory,
	},
}

var cmdAdd = &core.CommandStatic{
	Names: []string{
		"add",
		"new",
	},
	Description: "Add a command.",
	UsageArgs:   "<trigger> <text>",
	Run:         runAdd,
}

var cmdDel = &core.CommandStatic{
	Names: []string{
		"del",
		"delete",
		"rm",
		"remove",
	},
	Description: "Delete a command.",
	UsageArgs:   "<trigger>",
	Run:         runDel,
}

var cmdModify = &core.CommandStatic{
	Names: []string{
		"modify",
		"change",
	},
	Description: "Modify a command.",
	UsageArgs:   "<trigger> <text>",
	Run:         runModify,
}

var cmdList = &core.CommandStatic{
	Names: []string{
		"ls",
		"list",
	},
	Description: "List commands.",
	UsageArgs:   "",
	Run:         runList,
}

var cmdHistory = &core.CommandStatic{
	Names: []string{
		"history",
	},
	Description: "View a command's entire history of changes.",
	UsageArgs:   "<trigger>",
	Run:         runHistory,
}
