package nick

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Admin = &core.CommandStatic{
	Names: []string{
		"nick",
	},
	Run: runAdmin,

	Children: core.Commands{
		{
			Names: []string{
				"get",
			},
			Run: runAdminGet,
		},
		{
			Names: []string{
				"set",
			},
			Run: runAdminSet,
		},
	},
}
