package god

import (
	"context"
	"errors"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"
	"github.com/google/uuid"
	gogpt "github.com/sashabaranov/go-gpt3"
)

var (
	ErrIntervalTooShort = errors.New("The given interval is too short.")
	ErrRedeemNotSet     = errors.New("the streak redeem has not been set")
)

// Talk returns GPT3's response to a prompt.
func Talk(prompt string) (string, error) {
	c := gogpt.NewClient(core.OpenAIKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 60,
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
	return core.DB.PlaceGet("cmd_god_reply_on", place).Bool()
}

// ReplyOnSet will set the value that determines whether auto-replying is on or
// off (true or false) in the specified place.
func ReplyOnSet(place int64, on bool) error {
	return core.DB.PlaceSet("cmd_god_reply_on", place, on)
}

// ReplyIntervalGet returns the duration object of the interval that is
// required for auto-replies in the specified place.
func ReplyIntervalGet(place int64) (time.Duration, error) {
	return core.DB.PlaceGet("cmd_god_reply_interval", place).Duration()
}

// ReplyIntervalSet sets the reply interval for the specified place. Returns
// ErrIntervalTooShort if dur is larger than the global minimum that is allowed.
func ReplyIntervalSet(place int64, dur time.Duration) (error, error) {
	if core.MinGodInterval > dur {
		return ErrIntervalTooShort, nil
	}
	return nil, core.DB.PlaceSet("cmd_god_reply_interval", place, int(dur.Seconds()))
}

// ReplyLastGet returns a time object of the when the last reply happened
// in the specified place.
func ReplyLastGet(place int64) (time.Time, error) {
	return core.DB.PlaceGet("cmd_god_reply_last", place).Time()
}

// ReplyLastSet sets the timestamp of the last reply fot the specified place.
// The passed when is set to UTC before extracting the timestamp.
func ReplyLastSet(place int64, when time.Time) error {
	timestamp := when.UTC().Unix()
	return core.DB.PlaceSet("cmd_god_reply_last", place, timestamp)
}

func RedeemSet(place int64, id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return core.DB.PlaceSet("cmd_god_redeem", place, u)
}

func RedeemGet(place int64) (uuid.UUID, error, error) {
	id, isNil, err := core.DB.PlaceGet("cmd_god_redeem", place).OptionalUUID()
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	if isNil {
		return uuid.UUID{}, ErrRedeemNotSet, nil
	}
	return id, nil, nil
}
