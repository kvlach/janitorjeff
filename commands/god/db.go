package god

//import "github.com/janitorjeff/jeff-bot/core"

const dbTablePlaceSettings = "CommandGodPlaceSettings"

const dbSchema = `
-- All settings must come with default values, as those are used when first
-- adding a new entry.
CREATE TABLE IF NOT EXISTS CommandGodPlaceSettings (
	place INTEGER PRIMARY KEY,
	reply_on BOOLEAN NOT NULL DEFAULT FALSE,
	reply_interval INTEGER NOT NULL DEFAULT 1800, -- in seconds
	reply_last INTEGER NOT NULL DEFAULT 0 -- unix timestamp of last reply
);
`
