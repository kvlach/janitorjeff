package time

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"time",
	},
	Description: "time stuff and things",
	UsageArgs:   "(now | convert | timezone)",
	Run:         runNormal,
	Init:        init_,

	Children: core.Commands{
		{
			Names: []string{
				"now",
			},
			Description: "get the current time",
			UsageArgs:   "",
			Run:         runNormalNow,
		},
		{
			Names: []string{
				"convert",
			},
			Description: "convert to a timezone",
			UsageArgs:   "(now|<timestamp>) <timezone>",
			Run:         runNormalConvert,
		},
		{
			Names: []string{
				"timezone",
				"tz",
			},
			Description: "Set your own personal timezone.",
			UsageArgs:   "(set | delete | get)",
			Run:         runNormalTimezone,

			Children: core.Commands{
				{
					Names: []string{
						"set",
					},
					Description: "specify your own timezone",
					UsageArgs:   "<timezone>",
					Run:         runNormalTimezoneSet,
				},
				{
					Names: []string{
						"del",
						"delete",
						"rm",
						"remove",
					},
					Description: "specify your own timezone",
					UsageArgs:   "",
					Run:         runNormalTimezoneDelete,
				},
				{
					Names: []string{
						"get",
					},
					Description: "see your timezone",
					UsageArgs:   "",
					Run:         runNormalTimezoneGet,
				},
			},
		},
	},
}
