package god

import (
	"context"
	"errors"
	"time"

	"github.com/janitorjeff/jeff-bot/core"

	gogpt "github.com/sashabaranov/go-gpt3"
)

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

var ErrIntervalTooShort = errors.New("The given interval is too short.")

// Talk returns GPT3's response to a prompt.
func Talk(prompt string) (string, error) {
	c := gogpt.NewClient(core.OpenAIKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 40,
		Prompt:    prompt,
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Text, nil
}

// ReplyOnGet returns whether auto-replying is on or off (true or false) in the
// specified place.
func ReplyOnGet(place int64) (bool, error) {
	ret, err := core.DB.PlaceSettingGet(dbTablePlaceSettings, "reply_on", place)
	return ret.(bool), err
}

// ReplyOnSet will set the value that determines whether auto-replying is on or
// off (true or false) in the specified place.
func ReplyOnSet(place int64, on bool) error {
	return core.DB.PlaceSettingSet(dbTablePlaceSettings, "reply_on", place, on)
}

// ReplyIntervalGet returns the duration object of the interval that is
// required for auto-replies in the specified place.
func ReplyIntervalGet(place int64) (time.Duration, error) {
	interval, err := core.DB.PlaceSettingGet(dbTablePlaceSettings, "reply_interval", place)
	if err != nil {
		return time.Second, err
	}
	return time.Duration(interval.(int64)) * time.Second, nil
}

// ReplyIntervalSet sets the reply interval for the specified place. Returns
// ErrIntervalTooShort if dur is larger than the global minimum that is allowed.
func ReplyIntervalSet(place int64, dur time.Duration) (error, error) {
	if core.MinGodInterval > dur {
		return ErrIntervalTooShort, nil
	}
	return nil, core.DB.PlaceSettingSet(dbTablePlaceSettings, "reply_interval", place, int(dur.Seconds()))
}

// ReplyLastGet returns a time object of the when the last reply happened
// in the specified place.
func ReplyLastGet(place int64) (time.Time, error) {
	last, err := core.DB.PlaceSettingGet(dbTablePlaceSettings, "reply_last", place)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(last.(int64), 0).UTC(), nil
}

// ReplyLastSet sets the timestamp of the last reply fot the specified place.
// The passed when is set to UTC before extracting the timestamp.
func ReplyLastSet(place int64, when time.Time) error {
	timestamp := when.UTC().Unix()
	return core.DB.PlaceSettingSet(dbTablePlaceSettings, "reply_last", place, timestamp)
}

// ShouldReply returns true if the required interval has passed since the last
// auto-reply happened, meaning that the bot should send a new reply again.
func ShouldReply(place int64) (bool, error) {
	interval, err := ReplyIntervalGet(place)
	if err != nil {
		return false, err
	}
	last, err := ReplyLastGet(place)
	if err != nil {
		return false, err
	}
	diff := time.Now().UTC().Sub(last)
	return diff > interval, nil
}
