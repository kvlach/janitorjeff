package god

import (
	"context"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

var (
	UrrIntervalTooShort = core.UrrNew("The given interval is too short.")
)

// Talk returns GPT3.5's response to a user prompt.
func Talk(prompt string) (string, error) {
	log.Debug().Str("prompt", prompt).Msg("talking to gpt3.5")

	client := openai.NewClient(core.OpenAIKey)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 80,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are God who has taken the form of a janitor. You are a bit of an asshole. Always respond in English. Never use slurs. Respond with 300 characters or less.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
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
// UrrIntervalTooShort if dur is larger than the global minimum that is allowed.
func ReplyIntervalSet(place int64, dur time.Duration) (error, error) {
	if core.MinGodInterval > dur {
		return UrrIntervalTooShort, nil
	}
	return nil, core.DB.PlaceSet("cmd_god_reply_interval", place, int(dur.Seconds()))
}

func RedeemSet(place int64, id string) error {
	u, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return core.DB.PlaceSet("cmd_god_redeem", place, u)
}

// RedeemGet returns the place's god triggering redeem.
// If no redeem has been set returns core.UrrValNil.
func RedeemGet(place int64) (uuid.UUID, core.Urr, error) {
	return core.DB.PlaceGet("cmd_god_redeem", place).UUIDNil()
}
