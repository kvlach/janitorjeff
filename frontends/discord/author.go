package discord

import (
	"errors"
	"sync"

	"github.com/kvlach/janitorjeff/core"

	dg "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func getAuthorScope(authorID string) (int64, error) {
	rdbKey := "frontend_discord_scope_author_" + authorID

	return core.CacheScope(rdbKey, func() (int64, error) {
		return dbGetPersonScope(authorID)
	})
}

// author implements the core.Personifier interface
type author struct {
	user    *dg.User
	member  *dg.Member
	guildID string

	mu    sync.Mutex
	scope int64
}

// NewAuthor initializes an author struct which implements core.Personifier.
//   - usr can't be nil and must contain a non-empty usr.ID
//   - If mem is unknown, pass nil; will be lazily fetched (requires a guild ID)
//   - If the guild ID is unknown or the author is in a DM, pass an empty string
//   - The guild ID can be passed either through guildID or mem.GuildID
func NewAuthor(usr *dg.User, mem *dg.Member, guildID string) (core.Personifier, error) {
	if usr == nil {
		return nil, errors.New("usr must not be nil")
	}
	if usr.ID == "" {
		return nil, errors.New("usr must have a non-empty user ID")
	}

	if mem != nil {
		if guildID == "" && mem.GuildID == "" {
			return nil, errors.New("non nil mem implies a guildID")
		}
		if guildID == "" {
			guildID = mem.GuildID
		}
	}

	// no `guildID == ""` check, since it *will* be empty in case of a DM

	return &author{
		user:    usr,
		member:  mem,
		guildID: guildID,
	}, nil
}

// fetchUser fetches and returns the user, by *always* making an API call.
// Updates the value in the author struct.
// Used for partially instantiated users.
// The caller is responsible for locking the mutex.
func (a *author) fetchUser() (*dg.User, error) {
	usr, err := Client.Session.User(a.member.User.ID)
	if err != nil {
		return nil, err
	}
	a.user = usr
	return usr, nil
}

// updateMember fetches, caches and returns a Member object.
// Requires non-empty guildID.
// The caller is responsible for locking the mutex.
func (a *author) updateMember() (*dg.Member, error) {
	// if DM or unknown
	if a.guildID == "" {
		return nil, errors.New("don't have a guild ID, can't find member")
	}

	mem, err := Client.Member(a.guildID, a.user.ID)
	if err != nil {
		return nil, err
	}
	a.member = mem
	return mem, nil
}

// name returns the username. The caller is responsible for locking the mutex.
func (a *author) name() (string, error) {
	// Could be empty in the case of a partially instantiated .user
	if a.user.Username != "" {
		return a.user.Username, nil
	}
	// In the off-chance that .member has it...
	if a.member != nil && a.member.User != nil && a.member.User.Username != "" {
		return a.member.User.Username, nil
	}

	// We don't currently have the username, so we need to make an API call
	usr, err := a.fetchUser()
	if err != nil {
		return "", err
	}
	return usr.Username, nil
}

func (a *author) id() (string, error) {
	return a.user.ID, nil
}

func (a *author) ID() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.id()
}

func (a *author) Name() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.name()
}

func (a *author) DisplayName() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.member != nil && a.member.Nick != "" {
		return a.member.Nick, nil
	}

	// As opposed to the username that is never empty, the display name could be
	// empty either because it was never initialized or because it has not been
	// set by the user on Discord. Meaning that we must assume the former
	// and always make an API call to confirm.

	// Can't fetch a member without a guild ID
	if a.guildID == "" {
		return a.name()
	}

	mem, err := a.updateMember()
	if err != nil {
		return "", err
	}

	// If the display name has been set
	if mem.Nick != "" {
		return mem.Nick, nil
	} else {
		return a.name()
	}
}

func (a *author) Mention() (string, error) {
	id, err := a.ID()
	if err != nil {
		return "", err
	}
	return "<@" + id + ">", nil
}

func (a *author) BotAdmin() (bool, error) {
	id, err := a.ID()
	if err != nil {
		return false, err
	}
	return isBotAdmin(id), nil
}

func (a *author) Admin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isAdmin(a.guildID, aid)
}

func (a *author) Moderator() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.guildID, aid)
}

func (a *author) Subscriber() (bool, error) {
	return false, nil
}

func (a *author) Scope() (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.scope != 0 {
		log.Debug().
			Int64("scope", a.scope).
			Msg("LOCAL: found cached scope")
		return a.scope, nil
	}

	aid, err := a.id()
	if err != nil {
		return 0, err
	}
	scope, err := getAuthorScope(aid)
	if err != nil {
		return 0, err
	}
	a.scope = scope
	return scope, nil
}

// AuthorInteraction implements the core.Personifier interface for interactions
type AuthorInteraction struct {
	GuildID string
	Member  *dg.Member
	User    *dg.User

	scope int64
}

func (a *AuthorInteraction) ID() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.ID != "" {
		return a.Member.User.ID, nil
	}
	if a.User != nil && a.User.ID != "" {
		return a.User.ID, nil
	}
	return "", errors.New("can't figure out author id")
}

func (a *AuthorInteraction) Name() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}
	if a.User != nil && a.User.Username != "" {
		return a.User.Username, nil
	}
	return "", errors.New("can't figure out author name")
}

func (a *AuthorInteraction) DisplayName() (string, error) {
	if a.Member != nil && a.Member.User != nil && a.Member.User.Username != "" {
		return a.Member.User.Username, nil
	}
	if a.User != nil && a.User.Username != "" {
		return a.User.Username, nil
	}
	return "", errors.New("can't figure out author display name")
}

func (a *AuthorInteraction) Mention() (string, error) {
	// Member.Mention() requires Member.User
	if a.Member != nil && a.Member.User != nil {
		return a.Member.Mention(), nil
	}
	if a.User != nil {
		return a.User.Mention(), nil
	}
	return "", errors.New("can't figure out author mention")
}

func (a *AuthorInteraction) BotAdmin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isBotAdmin(aid), nil
}

func (a *AuthorInteraction) Admin() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isAdmin(a.GuildID, aid)
}

func (a *AuthorInteraction) Moderator() (bool, error) {
	aid, err := a.ID()
	if err != nil {
		return false, err
	}
	return isMod(a.GuildID, aid)
}

func (a *AuthorInteraction) Subscriber() (bool, error) {
	return false, nil
}

func (a *AuthorInteraction) Scope() (int64, error) {
	if a.scope != 0 {
		return a.scope, nil
	}
	aid, err := a.ID()
	if err != nil {
		return 0, err
	}
	scope, err := getAuthorScope(aid)
	if err != nil {
		return 0, err
	}
	a.scope = scope
	return scope, nil
}
