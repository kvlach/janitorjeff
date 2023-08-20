package god

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

var (
	UrrIntervalTooShort    = core.UrrNew("The given interval is too short.")
	UrrPersonalityNotFound = core.UrrNew("The given personality could not be found.")
	UrrPersonalityExists   = core.UrrNew("This personality already exists, try editing instead.")
	UrrOneLeft             = core.UrrNew("Only one personality left, better not delete it.")
	UrrGlobalPersonality   = core.UrrNew("This is a global personality, can't delete.")
)

// Talk returns GPT3.5's response to a user prompt.
func Talk(userPrompt string, place int64) (string, error) {
	p, err := PersonalityActive(place)
	if err != nil {
		return "", err
	}

	log.Debug().
		Str("system-prompt", p.Prompt).
		Str("user-prompt", userPrompt).
		Int64("place", place).
		Msg("talking to gpt3.5")

	resp, err := openai.NewClient(core.OpenAIKey).CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 80,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: p.Prompt,
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

type Personality struct {
	ID     int64
	Name   string
	Prompt string
	Global bool
}

func personalityActive(tx *core.Tx, place int64) (Personality, error) {
	// Since the settings table is accessed, make sure the row is there
	if err := tx.PlaceEnsure(place); err != nil {
		return Personality{}, err
	}

	var name, prompt string
	var placeDB *int64
	var id int64

	err := tx.Tx.QueryRow(`
		SELECT cgp.id, cgp.name, cgp.prompt, cgp.place
		FROM settings_place sp
		INNER JOIN cmd_god_personalities cgp ON sp.cmd_god_personality = cgp.id
		WHERE sp.place = $1 AND (cgp.place = $1 OR cgp.place IS NULL);
    `, place).Scan(&id, &name, &prompt, &placeDB)
	if err != nil {
		return Personality{}, err
	}
	return Personality{
		ID:     id,
		Name:   name,
		Prompt: prompt,
		Global: placeDB == nil,
	}, nil
}

// PersonalityActive returns the currently selected personality for place.
// Assumes that at least one exists, as it probably does because of globals.
func PersonalityActive(place int64) (Personality, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return Personality{}, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	p, err := personalityActive(tx, place)
	if err != nil {
		return Personality{}, err
	}
	return p, tx.Commit()
}

func personalitySet(tx *core.Tx, place int64, name string) (string, core.Urr, error) {
	ps, err := personalitiesList(tx, place)
	if err != nil {
		return "", nil, err
	}

	name = strings.ToLower(name)

	exists := false
	var id int64
	for _, p := range ps {
		if p.Name == name {
			exists = true
			id = p.ID
			break
		}
	}
	if !exists {
		return name, UrrPersonalityNotFound, nil
	}
	return name, nil, tx.PlaceSet("cmd_god_personality", place, id)
}

// PersonalitySet updates the active personality in place.
// Returns UrrPersonalityNotFound if name doesn't correspond to one.
func PersonalitySet(place int64, name string) (string, core.Urr, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return "", nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	name, urr, err := personalitySet(tx, place, name)
	if err != nil {
		return "", nil, err
	}
	return name, urr, tx.Commit()
}

func personalityExists(tx *core.Tx, place int64, name string) (bool, error) {
	var exists bool
	err := tx.Tx.QueryRow(`
		SELECT EXISTS (
		    SELECT 1 FROM cmd_god_personalities
		    WHERE name = $1 AND (place = $2 OR place IS NULL)
		)
	`, name, place).Scan(&exists)
	return exists, err
}

func personalityGlobal(tx *core.Tx, place int64, name string) (bool, error) {
	var placeDB *int64
	err := tx.Tx.QueryRow(`
		SELECT place
		FROM cmd_god_personalities
		WHERE name = $1 AND (place = $2 OR place IS NULL)
	`, name, place).Scan(&placeDB)
	if err != nil {
		return false, err
	}
	return placeDB == nil, nil
}

// PersonalityAdd adds a new personality called name in place.
// Returns UrrPersonalityExists if one by that name already exists in place.
// To edit an existing personality's prompt use PersonalityEdit.
func PersonalityAdd(place int64, name, prompt string) (core.Urr, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	name = strings.ToLower(name)

	exists, err := personalityExists(tx, place, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return UrrPersonalityExists, nil
	}

	_, err = core.DB.DB.Exec(`
		INSERT INTO cmd_god_personalities (place, name, prompt)
		VALUES ($1, $2, $3)
	`, place, name, prompt)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("name", name).
		Str("prompt", prompt).
		Msg("added personality")

	return nil, err
}

// PersonalityEdit edits the specified personality in place.
// Returns UrrPersonalityNotFound if name doesn't correspond to any.
// Returns UrrGlobalPersonality if it's global and not place-defined.
func PersonalityEdit(place int64, name, prompt string) (core.Urr, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	name = strings.ToLower(name)

	exists, err := personalityExists(tx, place, name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return UrrPersonalityNotFound, nil
	}

	global, err := personalityGlobal(tx, place, name)
	if err != nil {
		return nil, err
	}
	if global {
		return UrrGlobalPersonality, nil
	}

	_, err = tx.Tx.Exec(`
		UPDATE cmd_god_personalities
		SET prompt = $3
		WHERE place = $1 AND name = $2
	`, place, name, prompt)

	log.Debug().
		Err(err).
		Int64("place", place).
		Str("name", name).
		Str("prompt", prompt).
		Msg("edited personality")

	if err != nil {
		return nil, err
	}
	return nil, tx.Commit()
}

// PersonalityDelete will delete the given personality in place.
// Returns UrrOneLeft if name is the only personality left.
// Returns UrrPersonalityNotFound if name doesn't correspond to any.
// Returns UrrGlobalPersonality if it's global and not place-defined.
func PersonalityDelete(place int64, name string) (core.Urr, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	name = strings.ToLower(name)

	ps, err := personalitiesList(tx, place)
	if err != nil {
		return nil, err
	}
	if len(ps) == 1 {
		return UrrOneLeft, nil
	}

	var p Personality

	exists := false
	for _, per := range ps {
		if per.Name == name {
			p = per
			exists = true
			break
		}
	}
	if !exists {
		return UrrPersonalityNotFound, nil
	}
	if p.Global {
		return UrrGlobalPersonality, nil
	}

	active, err := personalityActive(tx, place)
	if err != nil {
		return nil, err
	}
	if active.Name == name {
		if _, _, err := personalitySet(tx, place, "neutral"); err != nil {
			return nil, err
		}
	}

	_, err = tx.Tx.Exec(`
		DELETE FROM cmd_god_personalities
		WHERE place = $1 AND name = $2
	`, place, name)

	if err != nil {
		return nil, err
	}
	return nil, tx.Commit()
}

// PersonalityGet returns the specified personality in place.
// Returns UrrPersonalityNotFound if name doesn't correspond to any.
func PersonalityGet(place int64, name string) (Personality, core.Urr, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return Personality{}, nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	name = strings.ToLower(name)

	exists, err := personalityExists(tx, place, name)
	if err != nil {
		return Personality{}, nil, err
	}
	if !exists {
		return Personality{}, UrrPersonalityNotFound, nil
	}

	p := Personality{Name: name}
	err = tx.Tx.QueryRow(`
		SELECT id, prompt
		FROM cmd_god_personalities
		WHERE (place = $1 OR place IS NULL) AND name = $2
	`, place, name).Scan(&p.ID, &p.Prompt)
	if err != nil {
		return Personality{}, nil, err
	}
	return p, nil, tx.Commit()
}

func personalitiesList(tx *core.Tx, place int64) ([]Personality, error) {
	rows, err := tx.Tx.Query(`
		SELECT id, name, prompt, place
		FROM cmd_god_personalities
		WHERE place = $1 OR place IS NULL
	`, place)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}(rows)

	var ps []Personality
	for rows.Next() {
		var id int64
		var place *int64
		var name, prompt string
		if err := rows.Scan(&id, &name, &prompt, &place); err != nil {
			return nil, err
		}
		ps = append(ps, Personality{ID: id, Name: name, Prompt: prompt, Global: place == nil})
	}

	err = rows.Err()

	log.Debug().
		Err(err).
		Int64("place", place).
		Interface("personalities", ps).
		Msg("got personalities")

	return ps, err
}

// PersonalitiesList returns the list of all available personalities available
// in place, includes both global and place-defined ones.
func PersonalitiesList(place int64) ([]Personality, error) {
	tx, err := core.DB.Begin()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	ps, err := personalitiesList(tx, place)
	if err != nil {
		return nil, err
	}
	return ps, tx.Commit()
}
