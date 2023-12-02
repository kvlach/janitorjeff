package twitch

import (
	"errors"
	"fmt"
	"net/http"

	"git.sr.ht/~slowtyper/janitorjeff/core"

	"git.sr.ht/~slowtyper/gosafe"
	"github.com/nicklaw5/helix/v2"
	"github.com/rs/zerolog/log"
)

var (
	ErrRetry             = errors.New("refresh the access token and try again")
	ErrNoResults         = errors.New("couldn't find what you were looking for")
	ErrUserTokenRequired = errors.New("this channel's broadcaster must connect their twitch account to the bot")
)

var appAccessToken = gosafe.Value[string]{}

type Helix struct {
	c *helix.Client
}

func NewHelix(userID string) (*Helix, error) {
	hx, err := helix.NewClient(&helix.Options{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	})
	if err != nil {
		return nil, err
	}

	if at, rt, err := dbGetUserTokens(userID); err == nil {
		if at == "" || rt == "" {
			hx.SetAppAccessToken(appAccessToken.Get())
		} else {
			hx.SetUserAccessToken(at)
			hx.SetRefreshToken(rt)
			hx.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
				err := dbUpdateUserTokens(userID, newAccessToken, newRefreshToken)
				if err != nil {
					log.Error().
						Err(err).
						Msg("POSTGRES: failed to save new user tokens")
				}
			})
		}
	} else {
		hx.SetAppAccessToken(appAccessToken.Get())
	}

	return &Helix{hx}, nil
}

func generateAppAccessToken() error {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	})
	if err != nil {
		return err
	}

	resp, err := client.RequestAppAccessToken([]string{"user:read:email"})
	if err != nil {
		return err
	}

	appAccessToken.Set(resp.Data.AccessToken)
	return nil
}

func checkErrors(err error, resp helix.ResponseCommon, length int) error {
	if err != nil {
		return fmt.Errorf("helix error: %v", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrRetry
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%d, %s: %s", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	if length == 0 {
		return ErrNoResults
	}

	return nil
}

func (hx *Helix) newAppAccessToken() error {
	// can't refresh an app access token, just get a new one
	err := generateAppAccessToken()
	if err != nil {
		return err
	}
	hx.c.SetAppAccessToken(appAccessToken.Get())
	return nil
}

func (hx *Helix) refreshToken() error {
	// the wrapper library handles refreshing user access tokens
	if hx.c.GetUserAccessToken() != "" {
		return nil
	}
	return hx.newAppAccessToken()
}

func (hx *Helix) GetFollower(broadcasterID, userID string) (helix.UserFollow, error) {
	resp, err := hx.c.GetUsersFollows(&helix.UsersFollowsParams{
		FromID: userID,
		ToID:   broadcasterID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Follows))

	switch err {
	case nil:
		return resp.Data.Follows[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.UserFollow{}, err
		}
		return hx.GetFollower(broadcasterID, userID)

	default:
		return helix.UserFollow{}, err
	}
}

func (hx *Helix) GetStream(broadcasterID string) (helix.Stream, error) {
	resp, err := hx.c.GetStreams(&helix.StreamsParams{
		UserIDs: []string{broadcasterID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Streams))

	switch err {
	case nil:
		return resp.Data.Streams[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.Stream{}, err
		}
		return hx.GetStream(broadcasterID)

	default:
		return helix.Stream{}, err
	}
}

func (hx *Helix) GetUser(userID string) (helix.User, error) {
	resp, err := hx.c.GetUsers(&helix.UsersParams{
		IDs: []string{userID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Users))

	switch err {
	case nil:
		return resp.Data.Users[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.User{}, err
		}
		return hx.GetUser(userID)

	default:
		return helix.User{}, err
	}
}

func (hx *Helix) GetUserID(username string) (string, error) {
	resp, err := hx.c.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Users))

	switch err {
	case nil:
		return resp.Data.Users[0].ID, nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return "", err
		}
		return hx.GetUserID(username)

	default:
		return "", err
	}
}

func (hx *Helix) GetClip(clipID string) (helix.Clip, error) {
	resp, err := hx.c.GetClips(&helix.ClipsParams{
		IDs: []string{clipID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Clips))

	switch err {
	case nil:
		return resp.Data.Clips[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.Clip{}, err
		}
		return hx.GetClip(clipID)

	default:
		return helix.Clip{}, err
	}
}

func (hx *Helix) GetBannedUser(broadcasterID, userID string) (helix.Ban, error) {
	resp, err := hx.c.GetBannedUsers(&helix.BannedUsersParams{
		BroadcasterID: broadcasterID,
		UserID:        userID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Bans))

	switch err {
	case nil:
		return resp.Data.Bans[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.Ban{}, err
		}
		return hx.GetBannedUser(broadcasterID, userID)

	default:
		return helix.Ban{}, err
	}
}

// func (h *Helix) IsUserBanned(broadcasterID, userID string) (bool, error) {
// 	resp, err := h.c.GetBannedUsers(&helix.BannedUsersParams{
// 		BroadcasterID: broadcasterID,
// 		UserID:        userID,
// 	})

// 	if err := checkErrors(err, resp.ResponseCommon, 1); err != nil {
// 		return false, fmt.Errorf("failed to get banned user '%s' in channel '%s': %v", userID, broadcasterID, err)
// 	}

// 	if len(resp.Data.Bans) > 1 {
// 		return true, nil
// 	}
// 	return false, nil
// }

func (hx *Helix) SearchGame(gameName string) (helix.Game, error) {
	resp, err := hx.c.GetGames(&helix.GamesParams{
		Names: []string{gameName},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Games))

	switch err {
	case nil:
		return resp.Data.Games[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.Game{}, err
		}
		return hx.SearchGame(gameName)

	default:
		return helix.Game{}, err
	}
}

func (hx *Helix) GetChannelInfo(broadcasterID string) (helix.ChannelInformation, error) {
	resp, err := hx.c.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{broadcasterID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Channels))

	switch err {
	case nil:
		return resp.Data.Channels[0], nil

	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return helix.ChannelInformation{}, err
		}
		return hx.GetChannelInfo(broadcasterID)

	default:
		return helix.ChannelInformation{}, err
	}
}

func (hx *Helix) GetGameName(channelId string) (string, error) {
	ch, err := hx.GetChannelInfo(channelId)
	return ch.GameName, err
}

func (hx *Helix) GetTitle(channelId string) (string, error) {
	ch, err := hx.GetChannelInfo(channelId)
	return ch.Title, err
}

func (hx *Helix) EditChannelInfo(broadcasterID, title, gameID string) (error, error) {
	if hx.c.GetUserAccessToken() == "" {
		return ErrUserTokenRequired, nil
	}

	// both the title and the game need to be set at the same time
	resp, err := hx.c.EditChannelInformation(&helix.EditChannelInformationParams{
		BroadcasterID: broadcasterID,
		Title:         title,
		GameID:        gameID,
	})

	err = checkErrors(err, resp.ResponseCommon, 1)

	switch err {
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return nil, err
		}
		return hx.EditChannelInfo(broadcasterID, title, gameID)

	default:
		return nil, err
	}
}

func (hx *Helix) SetTitle(channelID, title string) (error, error) {
	ch, err := hx.GetChannelInfo(channelID)
	if err != nil {
		return nil, err
	}
	return hx.EditChannelInfo(channelID, title, ch.GameID)
}

func (hx *Helix) SetGame(channelID, gameName string) (string, error, error) {
	title, err := hx.GetTitle(channelID)
	if err != nil {
		return "", nil, err
	}

	// Clears the game
	if gameName == "-" {
		urr, err := hx.EditChannelInfo(channelID, title, "0")
		return "nothing", urr, err
	}

	g, err := hx.SearchGame(gameName)
	if err != nil {
		return "", nil, err
	}
	urr, err := hx.EditChannelInfo(channelID, title, g.ID)
	return g.Name, urr, err
}

func (hx *Helix) RedeemsList(broadcasterID string) ([]helix.ChannelCustomReward, error) {
	resp, err := hx.c.GetCustomRewards(&helix.GetCustomRewardsParams{
		BroadcasterID: broadcasterID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.ChannelCustomRewards))

	switch err {
	case nil:
		log.Debug().
			Interface("redeems", resp.Data.ChannelCustomRewards).
			Msg("got redeems")
		return resp.Data.ChannelCustomRewards, nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return nil, err
		}
		return hx.RedeemsList(broadcasterID)
	default:
		return nil, err
	}
}

func (hx *Helix) CreateSubscription(broadcasterID, t string) (string, error) {
	resp, err := hx.c.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    t,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://" + core.VirtualHost + CallbackEventSub,
			Secret:   secret,
		},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.EventSubSubscriptions))

	switch err {
	case nil:
		id := resp.Data.EventSubSubscriptions[0].ID
		log.Debug().Str("type", t).Str("id", id).Msg("created subscription")
		return id, nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return "", err
		}
		return hx.CreateSubscription(broadcasterID, t)
	default:
		return "", err
	}
}

func (hx *Helix) DeleteSubscription(subID string) error {
	resp, err := hx.c.RemoveEventSubSubscription(subID)
	if err != nil {
		return err
	}

	err = checkErrors(err, resp.ResponseCommon, 1)

	switch err {
	case nil:
		return nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return err
		}
		return hx.DeleteSubscription(subID)
	default:
		return err
	}
}

func (hx *Helix) ListSubscriptions() ([]helix.EventSubSubscription, error) {
	resp, err := hx.c.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{})
	if err != nil {
		return nil, err
	}

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.EventSubSubscriptions))

	switch err {
	case nil:
		return resp.Data.EventSubSubscriptions, nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return nil, err
		}
		return hx.ListSubscriptions()
	default:
		return nil, err
	}
}

func (hx *Helix) IsMod(broadcasterID, userID string) (bool, error) {
	resp, err := hx.c.GetModerators(&helix.GetModeratorsParams{
		BroadcasterID: broadcasterID,
		UserIDs:       []string{userID},
	})

	switch err := checkErrors(err, resp.ResponseCommon, 1); err {
	case nil:
		return len(resp.Data.Moderators) > 1, nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return false, err
		}
		return hx.IsMod(broadcasterID, userID)
	default:
		return false, err
	}
}

func (hx *Helix) IsSub(broadcasterID, userID string) (bool, error) {
	resp, err := hx.c.CheckUserSubscription(&helix.UserSubscriptionsParams{
		BroadcasterID: broadcasterID,
		UserID:        userID,
	})

	switch err := checkErrors(err, resp.ResponseCommon, 1); err {
	case nil:
		return len(resp.Data.UserSubscriptions) > 1, nil
	case ErrRetry:
		if err := hx.refreshToken(); err != nil {
			return false, err
		}
		return hx.IsSub(broadcasterID, userID)
	default:
		return false, err
	}
}
