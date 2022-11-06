package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"prefix",
	},
	Description: "",
	UsageArgs:   "",
	Frontends:   frontends.All,
	Run:         runAdmin,
	Children: core.Commands{
		cmdAdminAdd,
		cmdAdminDel,
		cmdAdminList,
		cmdAdminReset,
	},
}

var cmdAdminAdd = &core.CommandStatic{
	Names:       core.Add,
	Description: "add prefix",
	UsageArgs:   "",
	Run:         runAdminAdd,
}

var cmdAdminDel = &core.CommandStatic{
	Names:       core.Delete,
	Description: "add prefix",
	UsageArgs:   "",
	Run:         runAdminDel,
}

var cmdAdminList = &core.CommandStatic{
	Names:       core.List,
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
