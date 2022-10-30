package command

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Command = &core.CommandStatic{
	Names: []string{
		"cmd",
		"command",
	},
	Description: "Add, edit, delete or list custom commands.",
	UsageArgs:   "(add | edit | delete | list)",
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
	Names:       core.Add,
	Description: "Add a command.",
	UsageArgs:   "<trigger> <text>",
	Run:         runAdd,
}

var cmdDel = &core.CommandStatic{
	Names:       core.Delete,
	Description: "Delete a command.",
	UsageArgs:   "<trigger>",
	Run:         runDel,
}

var cmdModify = &core.CommandStatic{
	Names:       core.Edit,
	Description: "Edit a command.",
	UsageArgs:   "<trigger> <text>",
	Run:         runEdit,
}

var cmdList = &core.CommandStatic{
	Names:       core.List,
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
