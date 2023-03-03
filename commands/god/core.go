package god

import (
	"context"

	"github.com/janitorjeff/jeff-bot/core"

	gogpt "github.com/sashabaranov/go-gpt3"
)

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
