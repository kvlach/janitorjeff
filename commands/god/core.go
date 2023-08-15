package god

import (
	"context"
	"fmt"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

var (
	UrrIntervalTooShort = core.UrrNew("The given interval is too short.")
)

type Mood int

const (
	MoodDefault Mood = iota
	MoodRude
	MoodSad
)

func (m Mood) String() string {
	switch m {
	case MoodDefault:
		return "Default"
	case MoodRude:
		return "Rude"
	case MoodSad:
		return "Sad"
	default:
		panic("unknown mood")
	}
}

func SystemPrompt(mood Mood) (string, error) {
	switch mood {
	case MoodDefault:
		return "You are God who has taken the form of a janitor. You are a bit of an asshole, but not too much. You are goofy. Always respond in English. Respond with 300 characters or less.", nil
	case MoodRude:
		return "You are God who has taken the form of a janitor. You are rude. Always respond in English. Respond with 300 characters or less.", nil
	case MoodSad:
		return "You are God who has taken the form of a janitor. You are very sad about everything. Respond in 300 characters or less.", nil
	default:
		return "", fmt.Errorf("unknown mood '%d'", mood)
	}
}

// Talk returns GPT3.5's response to a user prompt.
func Talk(userPrompt string, place int64) (string, error) {
	mood, err := MoodGet(place)
	if err != nil {
		return "", err
	}
	systemPrompt, err := SystemPrompt(mood)
	if err != nil {
		return "", err
	}

	log.Debug().Str("user-prompt", userPrompt).Msg("talking to gpt3.5")
	resp, err := openai.NewClient(core.OpenAIKey).CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 80,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
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

func MoodSet(place int64, mood Mood) error {
	return core.DB.PlaceSet("cmd_god_mood", place, int(mood))
}

func MoodGet(place int64) (Mood, error) {
	mood, err := core.DB.PlaceGet("cmd_god_mood", place).Int()
	if err != nil {
		return 0, err
	}
	return Mood(mood), nil
}
