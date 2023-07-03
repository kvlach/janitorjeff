package twitch

import (
	"errors"

	"git.sr.ht/~slowtyper/janitorjeff/core"
)

// Author implements both the core.Author and core.Here interfaces since users
// and channels are the same thing on twitch.
type Author struct {
	id          string
	username    string
	displayName string
	badges      map[string]int
	roomID      string
}

func (a Author) ID() (string, error) {
	if a.id != "" {
		return a.id, nil
	}

	// doesn't call a.Name() on purpose as that method tries to use the id to
	// figure out the username
	if a.username == "" {
		return "", errors.New("no username provided, can't figure out the id")
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

func (a Author) Name() (string, error) {
	if a.username != "" {
		return a.username, nil
	}

	if a.id == "" {
		return "", errors.New("no id provided, can't figure out the username")
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
	return a.username, nil
}

func (a Author) DisplayName() (string, error) {
	if a.displayName != "" {
		return a.displayName, nil
	}

	if a.id == "" {
		return "", errors.New("no id provided, can't figure out the username")
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
	return a.displayName, nil
}

func (a Author) Mention() (string, error) {
	n, err := a.DisplayName()
	if err != nil {
		return "", err
	}
	return "@" + n, nil
}

func (a Author) BotAdmin() (bool, error) {
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

func (a Author) Admin() (bool, error) {
	if a.badges != nil {
		_, ok := a.badges["broadcaster"]
		return ok, nil
	}
	if a.roomID == "" {
		return false, errors.New("no room id provided, can't figure out admin perms")
	}
	uid, err := a.ID()
	if err != nil {
		return false, err
	}
	return uid == a.roomID, nil
}

func (a Author) Moderator() (bool, error) {
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

	if a.roomID == "" {
		return false, errors.New("no broadcaster id provided, can't figure out mod perms")
	}

	h, err := Frontend.Helix()
	if err != nil {
		return false, err
	}
	uid, err := a.ID()
	if err != nil {
		return false, err
	}
	return h.IsMod(a.roomID, uid)
}

func (a Author) Subscriber() (bool, error) {
	if a.badges != nil {
		if _, ok := a.badges["subscriber"]; ok {
			return true, nil
		}
		_, ok := a.badges["founder"]
		return ok, nil
	}

	if a.roomID == "" {
		return false, errors.New("no broadcaster id provided, can't figure out sub status")
	}
	h, err := Frontend.Helix()
	if err != nil {
		return false, err
	}
	uid, err := a.ID()
	if err != nil {
		return false, err
	}
	return h.IsSub(a.roomID, uid)
}

func (a Author) Scope() (int64, error) {
	id, err := a.ID()
	if err != nil {
		return 0, err
	}
	name, err := a.Name()
	if err != nil {
		return 0, err
	}
	return core.CacheScope("frontend_twitch_scope_"+id, func() (int64, error) {
		return dbAddChannelSimple(id, name)
	})
}
