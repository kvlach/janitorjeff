package twitch

import (
	"errors"

	"git.sr.ht/~slowtyper/janitorjeff/core"
)

// author implements the core.Personifier interface.
type author struct {
	id          string
	username    string
	displayName string
	badges      map[string]int
	roomID      string

	scope int64
}

// NewAuthor initializes an author object.
// If id is unknown, pass an empty string.
// If username is unknown, pass an empty string.
// At least one of id or username must not be an empty string.
// If displayName is unknown, pass an empty string.
// If badges are unknown, pass nil.
// The roomID must not be an empty string.
func NewAuthor(id, username, displayName, roomID string, badges map[string]int) (core.Personifier, error) {
	if id == "" && username == "" {
		return nil, errors.New("at least one of id or username is required")
	}
	if roomID == "" {
		return nil, errors.New("roomID is required to be non-empty")
	}
	return author{
		id:          id,
		username:    username,
		displayName: displayName,
		badges:      badges,
		roomID:      roomID,
	}, nil
}

func (a author) ID() (string, error) {
	if a.id != "" {
		return a.id, nil
	}

	h, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	id, err := h.GetUserID(a.username)
	if err != nil {
		return "", err
	}
	a.id = id
	return id, nil
}

func (a author) Name() (string, error) {
	if a.username != "" {
		return a.username, nil
	}

	h, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	user, err := h.GetUser(a.id)
	if err != nil {
		return "", err
	}
	a.username = user.Login
	// might as well make sure the display name is saved
	a.displayName = user.DisplayName
	return a.username, nil
}

func (a author) DisplayName() (string, error) {
	if a.displayName != "" {
		return a.displayName, nil
	}

	h, err := Frontend.Helix()
	if err != nil {
		return "", err
	}
	user, err := h.GetUser(a.id)
	if err != nil {
		return "", err
	}
	a.displayName = user.DisplayName
	// might as well make sure the username is saved
	a.username = user.Login
	return a.displayName, nil
}

func (a author) Mention() (string, error) {
	n, err := a.DisplayName()
	if err != nil {
		return "", err
	}
	return "@" + n, nil
}

func (a author) BotAdmin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	for _, admin := range Admins {
		if aid == admin {
			return true, nil
		}
	}
	return false, nil
}

func (a author) Admin() (bool, error) {
	if a.badges != nil {
		_, ok := a.badges["broadcaster"]
		return ok, nil
	}
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return aid == a.roomID, nil
}

func (a author) Moderator() (bool, error) {
	admin, err := a.Admin()
	if err != nil {
		return false, err
	}
	if admin {
		return true, nil
	}

	if a.badges != nil {
		_, ok := a.badges["moderator"]
		return ok, nil
	}

	h, err := Frontend.Helix()
	if err != nil {
		return false, err
	}
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return h.IsMod(a.roomID, aid)
}

func (a author) Subscriber() (bool, error) {
	if a.badges != nil {
		if _, ok := a.badges["subscriber"]; ok {
			return true, nil
		}
		_, ok := a.badges["founder"]
		return ok, nil
	}

	h, err := Frontend.Helix()
	if err != nil {
		return false, err
	}
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return h.IsSub(a.roomID, aid)
}

func (a author) Scope() (int64, error) {
	if a.scope != 0 {
		return a.scope, nil
	}
	aid, err := a.ID()
	if err != nil {
		return 0, err
	}
	scope, err := core.CacheScope("frontend_twitch_scope_"+aid, func() (int64, error) {
		return dbAddChannel(aid)
	})
	if err != nil {
		return 0, err
	}
	a.scope = scope
	return scope, nil
}
