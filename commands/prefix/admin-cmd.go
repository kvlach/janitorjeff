package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"prefix",
	},
	Description: "",
	UsageArgs:   "",
	Run:         runAdmin,
	Children: core.Commands{
		cmdAdminAdd,
		cmdAdminDel,
		cmdAdminList,
		cmdAdminReset,
	},
}

var cmdAdminAdd = &core.CommandStatic{
	Names: []string{
		"add",
		"new",
	},
	Description: "add prefix",
	UsageArgs:   "",
	Run:         runAdminAdd,
}

var cmdAdminDel = &core.CommandStatic{
	Names: []string{
		"del",
		"delete",
		"rm",
		"remove",
	},
	Description: "add prefix",
	UsageArgs:   "",
	Run:         runAdminDel,
}

var cmdAdminList = &core.CommandStatic{
	Names: []string{
		"ls",
		"list",
	},
	Description: "list prefixes",
	UsageArgs:   "",
	Run:         runAdminList,
}

var cmdAdminReset = &core.CommandStatic{
	Names: []string{
		"reset",
	},
	Description: "reset prefixes",
	UsageArgs:   "",
	Run:         runAdminReset,
}
