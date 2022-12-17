package discord

// Here implements the core.Here interface.
type Here struct {
	// cache scopes
	scopeExact   int64
	scopeLogical int64

	ChannelID string
	GuildID   string
}

func (h *Here) ID() string {
	return h.ChannelID
}

func (h *Here) Name() string {
	return h.ChannelID
}

func (h *Here) ScopeExact() (int64, error) {
	if h.scopeExact != 0 {
		return h.scopeExact, nil
	}
	here, err := getPlaceExactScope(h.ID(), h.ChannelID, h.GuildID)
	if err != nil {
		return -1, err
	}
	h.scopeExact = here
	return here, nil
}

func (h *Here) ScopeLogical() (int64, error) {
	if h.scopeLogical != 0 {
		return h.scopeLogical, nil
	}
	here, err := getPlaceLogicalScope(h.ID(), h.ChannelID, h.GuildID)
	if err != nil {
		return -1, err
	}
	h.scopeLogical = here
	return here, nil
}
