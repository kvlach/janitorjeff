package prefix

import (
	"github.com/kvlach/janitorjeff/core"
)

type flags struct {
	fs *core.Flags

	typeFlag  core.CommandType
	scopeFlag int64
}

func newFlags(m *core.EventMessage) *flags {
	f := &flags{
		fs: core.NewFlags(m),
	}
	return f
}

func (f *flags) TypeFlag() *flags {
	core.TypeFlag(&f.typeFlag, core.All, f.fs)
	return f
}

func (f *flags) ScopeFlag() *flags {
	core.FlagPlace(&f.scopeFlag, f.fs)
	return f
}
