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

func Talk(prompt string) (string, error) {
	c := gogpt.NewClient(core.OpenAIKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 20,
		Prompt:    prompt,
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Text, nil
}

func ReplyOnGet(place int64) (bool, error) {
	ret, err := core.DB.PlaceSettingsGet(dbTablePlaceSettings, "reply_on", place)
	return ret.(bool), err
}

func ReplyOnSet(place int64, on bool) error {
	return core.DB.PlaceSettingsSet(dbTablePlaceSettings, "reply_on", place, on)
}

func ReplyIntervalGet(place int64) (time.Duration, error) {
	interval, err := core.DB.PlaceSettingsGet(dbTablePlaceSettings, "reply_interval", place)
	if err != nil {
		return time.Second, err
	}
	return time.Duration(interval.(int64)) * time.Second, nil
}

func ReplyIntervalSet(place int64, dur time.Duration) (error, error) {
	if core.MinGodInterval > dur {
		return ErrIntervalTooShort, nil
	}
	return nil, core.DB.PlaceSettingsSet(dbTablePlaceSettings, "reply_interval", place, int(dur.Seconds()))
}

func ReplyLastGet(place int64) (time.Time, error) {
	last, err := core.DB.PlaceSettingsGet(dbTablePlaceSettings, "reply_last", place)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(last.(int64), 0).UTC(), nil
}

func ReplyLastSet(place int64, when time.Time) error {
	timestamp := when.UTC().Unix()
	return core.DB.PlaceSettingsSet(dbTablePlaceSettings, "reply_last", place, timestamp)
}

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
