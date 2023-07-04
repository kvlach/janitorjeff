package core

// Author is the interface used to abstract a frontend's message author.
type Author interface {
	// ID returns the author's ID, this should be a unique, static, identifier
	// in that frontend.
	ID() (string, error)

	// Name returns the author's username.
	Name() (string, error)

	// DisplayName return's the author's display name. If only usernames exist
	// for that frontend, then return the username.
	DisplayName() (string, error)

	// Mention return's a string that mention's the author. This should ideally
	// ping them in some way.
	Mention() (string, error)

	// BotAdmin returns true if the author is a bot admin, otherwise returns
	// false.
	BotAdmin() (bool, error)

	// Admin checks if the author is considered an admin. Should return true
	// only if the author has basically every permission.
	Admin() (bool, error)

	// Moderator checks if the author is considered a moderator. General rule of
	// thumb is that if the author can ban people, then they are mods.
	Moderator() (bool, error)

	// Subscriber returns true if the author is considered a subscriber. General
	// rule of thumb is that if they are paying money in some way, then they are
	// a subscriber. If no such thing exists for the specific frontend, then it
	// should always return false.
	Subscriber() (bool, error)

	// Scope return's the author's scope.
	// If it doesn't exist, it will create it and add it to the database.
	Scope() (author int64, err error)
}
