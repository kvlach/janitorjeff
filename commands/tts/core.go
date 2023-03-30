package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/janitorjeff/jeff-bot/core"
	"github.com/janitorjeff/jeff-bot/frontends/twitch"

	"github.com/janitorjeff/gosafe"
	"github.com/rs/zerolog/log"
)

var Hooks = gosafe.Map[string, int]{}

var (
	ErrHookNotFound   = errors.New("Wasn't monitoring, what are you even trynna do??")
	ErrPersonNotFound = errors.New("Person's voice has not been set.")
)

var Voices = []string{
	// DISNEY VOICES
	"en_us_ghostface",       // Ghost Face
	"en_us_chewbacca",       // Chewbacca
	"en_us_c3po",            // C3PO
	"en_us_stitch",          // Stitch
	"en_us_stormtrooper",    // Stormtrooper
	"en_us_rocket",          // Rocket
	"en_female_madam_leota", // Madame Leota
	"en_male_ghosthost",     // Ghost Host
	"en_male_pirate",        // Pirate

	// ENGLISH VOICES
	"en_au_001", // English AU - Female
	"en_au_002", // English AU - Male
	"en_uk_001", // English UK - Male 1
	"en_uk_003", // English UK - Male 2
	"en_us_001", // English US - Female 1
	"en_us_002", // English US - Female 2
	"en_us_006", // English US - Male 1
	"en_us_007", // English US - Male 2
	"en_us_009", // English US - Male 3
	"en_us_010", // English US - Male 4

	// EUROPE VOICES
	"fr_001", // French - Male 1
	"fr_002", // French - Male 2
	"de_001", // German - Female
	"de_002", // German - Male
	"es_002", // Spanish - Male

	// AMERICA VOICES
	"es_mx_002", // Spanish MX - Male
	"br_001",    // Portuguese BR - Female 1
	"br_003",    // Portuguese BR - Female 2
	"br_004",    // Portuguese BR - Female 3
	"br_005",    // Portuguese BR - Male

	// ASIA VOICES
	"id_001", // Indonesian - Female
	"jp_001", // Japanese - Female 1
	"jp_003", // Japanese - Female 2
	"jp_005", // Japanese - Female 3
	"jp_006", // Japanese - Male
	"kr_002", // Korean - Male 1
	"kr_003", // Korean - Female
	"kr_004", // Korean - Male 2

	// SINGING VOICES
	// "en_female_f08_salut_damour",       // Alto
	// "en_male_m03_lobby",                // Tenor
	// "en_male_m03_sunshine_soon",        // Sunshine Soon
	// "en_female_f08_warmy_breeze",       // Warmy Breeze
	// "en_female_ht_f08_glorious",        // Glorious
	// "en_male_sing_funny_it_goes_up",    // It Goes Up
	// "en_male_m2_xhxs_m03_silly",        // Chipmunk
	// "en_female_ht_f08_wonderful_world", // Dramatic

	// OTHER
	"en_male_narration",   // Narrator
	"en_male_funny",       // Wacky
	"en_female_emotional", // Peaceful
	"en_male_cody",        // Serious
}

type TTSResp struct {
	Data struct {
		SKey     string `json:"s_key"`
		VStr     string `json:"v_str"`
		Duration string `json:"duration"`
		Speaker  string `json:"speaker"`
	} `json:"data"`
	Extra struct {
		LogID string `json:"log_id"`
	} `json:"extra"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// TTS will return a slice of bytes containing the audio generated by the TikTok
// TTS. You need to have the TikTokSessionID global set.
func TTS(voice, text string) ([]byte, error) {
	reqURL := "https://api16-normal-useast5.us.tiktokv.com/media/api/text/speech/invoke/?"
	reqURL += "text_speaker=" + voice
	reqURL += "&req_text=" + url.QueryEscape(text)
	reqURL += "&speaker_map_type=0&aid=1233"

	client := &http.Client{}
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Cookie": {"sessionid=" + core.TikTokSessionID},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data TTSResp
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(data.Data.VStr)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// Play will, if necessary join the appropriate voice channel, and start playing
// the TTS specified by text.
func Play(sp core.AudioSpeaker, voice, text string) error {
	audio, err := TTS(voice, text)
	if err != nil {
		return err
	}

	err = sp.Join()
	if err != nil {
		return err
	}

	state := &core.AudioState{}
	state.Set(core.AudioPlay)

	buf := ioutil.NopCloser(bytes.NewReader(audio))
	core.AudioFFmpegBufferPipe(sp, buf, state)

	return nil
}

// Start will create a hook and will monitor all incoming messages, if they
// are from twitch and match the specified username then the the TTS audio will
// be sent to the specified speaker.
func Start(sp core.AudioSpeaker, twitchUsername string) {
	id := core.Hooks.Register(func(m *core.Message) {
		if m.Frontend.Type() != twitch.Frontend.Type() || m.Here.Name() != twitchUsername {
			return
		}

		here, err := m.Here.ScopeLogical()
		if err != nil {
			return
		}

		subonly, err := PlaceSubOnlyGet(here)
		if err != nil {
			return
		}
		if subonly && !(m.Author.Subscriber() || m.Author.Mod()) {
			log.Debug().Msg("author is neither a sub nor a mod, skipping")
			return
		}

		con, err := sp.AuthorConnected()
		log.Debug().
			Err(err).
			Bool("connected", con).
			Msg("checked if author is still connected")

		if err != nil {
			return
		}
		if !con {
			Stop(twitchUsername)
			return
		}

		deaf, err := sp.AuthorDeafened()

		log.Debug().
			Err(err).
			Bool("deaf", deaf).
			Msg("checked if author is deafened")

		if err != nil || deaf {
			return
		}

		for _, arg := range strings.Fields(m.Raw) {
			if core.IsValidURL(arg) {
				return
			}
		}

		author, err := m.Author.Scope()
		if err != nil {
			return
		}

		voice, err := PersonVoiceGet(author, here)
		if err != nil {
			return
		}

		Play(sp, voice, m.Raw)
	})
	Hooks.Set(twitchUsername, id)
}

// Stop will delete the hook created by Start. Returns ErrHookNotFound if the
// hook doesn't exist.
func Stop(twitchUsername string) error {
	id, ok := Hooks.Get(twitchUsername)
	if !ok {
		return ErrHookNotFound
	}
	core.Hooks.Delete(id)
	Hooks.Delete(twitchUsername)
	return nil
}

// PersonVoiceGet returns the person's voice in this place. If no voice has
// been set then it picks a random one and saves it.
func PersonVoiceGet(person, place int64) (string, error) {
	voice, err := core.DB.SettingPersonGet("cmd_tts_voice", person, place)
	if err != nil {
		return "", err
	}
	return voice.(string), nil
}

// PersonVoiceSet sets the user voice.
func PersonVoiceSet(person, place int64, voice string) error {
	exists := false
	for _, v := range Voices {
		if v == voice {
			exists = true
			break
		}
	}
	if !exists {
		return errors.New("invalid voice")
	}
	return core.DB.SettingPersonSet("cmd_tts_voice", person, place, voice)
}

// PlaceSubOnlyGet returns the sub-only state for the specified place.
func PlaceSubOnlyGet(place int64) (bool, error) {
	subonly, err := core.DB.SettingPlaceGet("cmd_tts_subonly", place)
	if err != nil {
		return false, err
	}
	return subonly.(bool), nil
}

// PlaceSubOnlySet sets the sub-only state for the specified place.
func PlaceSubOnlySet(place int64, state bool) error {
	return core.DB.SettingPlaceSet("cmd_tts_subonly", place, state)
}
