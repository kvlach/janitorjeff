package twitch

type Here struct {
	RoomID   string
	RoomName string
}

func (h Here) ID() string {
	return h.RoomID
}

func (h Here) Name() string {
	return h.RoomName
}

func (h Here) Scope() (int64, error) {
	return dbAddChannelSimple(h.ID(), h.Name())
}

func (h Here) ScopeExact() (int64, error) {
	return h.Scope()
}

func (h Here) ScopeLogical() (int64, error) {
	return h.Scope()
}
