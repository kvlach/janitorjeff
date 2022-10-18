package prefix

import (
	"git.slowtyper.com/slowtyper/janitorjeff/core"
)

type flags struct {
	fs *core.Flags

	typeFlag  int
	scopeFlag int64
}

func (f *flags) TypeFlag() *flags {
	core.TypeFlag(&f.typeFlag, core.All, f.fs)
	return f
}

func (f *flags) ScopeFlag() *flags {
	core.ScopeFlag(&f.scopeFlag, f.fs)
	return f
}
