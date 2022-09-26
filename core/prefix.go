package core

func (g *GlobalVars) Prefixes() []string {
	// Creates a copy instead of a reference to make modifying the prefixes,
	// e.g. for renering easier, since otherwise it would modify the global
	// variable.
	prefixes := make([]string, len(g.Prefixes_))
	copy(prefixes, g.Prefixes_)
	return prefixes
}
