package connect

import (
	"fmt"

	"git.slowtyper.com/slowtyper/janitorjeff/core"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends"
	"git.slowtyper.com/slowtyper/janitorjeff/frontends/twitch"

	"github.com/nicklaw5/helix"
)

var Normal = normal{}

type normal struct{}

func (normal) Type() core.CommandType {
	return core.Normal
}

func (normal) Frontends() int {
	return frontends.All
}

func (normal) Permitted(*core.Message) bool {
	return true
}

func (normal) Names() []string {
	return []string{
		"connect",
	}
}

func (normal) Description() string {
	return "Connect one of your accounts to the bot."
}

func (normal) UsageArgs() string {
	return "(twitch)"
}

func (normal) Parent() core.CommandStatic {
	return nil
}

func (normal) Children() core.CommandsStatic {
	return core.CommandsStatic{
		NormalTwitch,
	}
}

func (normal) Init() error {
	return nil
}

func (normal) Run(m *core.Message) (any, error, error) {
	return m.Usage(), core.ErrMissingArgs, nil
}

////////////
//        //
// twitch //
//        //
////////////

var NormalTwitch = normalTwitch{}

type normalTwitch struct{}

func (c normalTwitch) Type() core.CommandType {
	return c.Parent().Type()
}

func (c normalTwitch) Frontends() int {
	return c.Parent().Frontends()
}

func (c normalTwitch) Permitted(m *core.Message) bool {
	return c.Parent().Permitted(m)
}

func (normalTwitch) Names() []string {
	return []string{
		"twitch",
	}
}

func (normalTwitch) Description() string {
	return "Connect your twitch account to the bot."
}

func (normalTwitch) UsageArgs() string {
	return ""
}

func (normalTwitch) Parent() core.CommandStatic {
	return Normal
}

func (normalTwitch) Children() core.CommandsStatic {
	return nil
}

func (normalTwitch) Init() error {
	return nil
}

func (c normalTwitch) Run(m *core.Message) (any, error, error) {
	switch m.Frontend {
	case frontends.Discord:
		return c.discord(m)
	default:
		return c.text(m)
	}
}

func (c normalTwitch) discord(m *core.Message) (string, error, error) {
	url, err := c.core(m)
	return fmt.Sprintf("<%s>", url), nil, err
}

func (c normalTwitch) text(m *core.Message) (string, error, error) {
	url, err := c.core(m)
	return url, nil, err
}

func (normalTwitch) core(m *core.Message) (string, error) {
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

	authURL := c.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		Scopes:       scopes,
		State:        state,
		ForceVerify:  true,
	})

	return authURL, nil
}
