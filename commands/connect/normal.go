package connect

import (
	"fmt"
	"time"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"

	"github.com/nicklaw5/helix"
)

var Normal = &core.CommandStatic{
	Names: []string{
		"connect",
	},
	Description: "Connect one of your accounts to the bot.",
	UsageArgs:   "(twitch)",
	Frontends:   frontends.All,
	Run:         normalRun,

	Children: core.Commands{
		{
			Names: []string{
				"twitch",
			},
			Description: "Connect your twitch account to the bot.",
			UsageArgs:   "",
			Run:         normalRunTwitch,
		},
	},
}

func normalRun(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////
//        //
// twitch //
//        //
////////////

func normalRunTwitch(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return normalRunTwitchDiscord(m)
	default:
		return normalRunTwitchText(m)
	}
}

func normalRunTwitchDiscord(m *core.Message) (string, error, error) {
	url, err := runTwitchCore(m)
	return fmt.Sprintf("<%s>", url), nil, err
}

func normalRunTwitchText(m *core.Message) (string, error, error) {
	url, err := runTwitchCore(m)
	return url, nil, err
}

func runTwitchCore(m *core.Message) (string, error) {
	clientID := twitch.ClientID
	redirectURI := fmt.Sprintf("http://%s/twitch/callback", core.Globals.Host)

	c, err := helix.NewClient(&helix.Options{
		ClientID:    clientID,
		RedirectURI: redirectURI,
	})

	if err != nil {
		return "", err
	}

	scopes := []string{
		"channel:manage:broadcast",
		"channel:moderate",
		"moderation:read",
	}

	state, err := twitch.GetState()
	if err != nil {
		return "", err
	}

	twitch.States.Add(state)
	go func() {
		time.Sleep(1 * time.Minute)
		twitch.States.Delete(state)
	}()

	authURL := c.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		Scopes:       scopes,
		State:        state,
		ForceVerify:  true,
	})

	return authURL, nil
}
