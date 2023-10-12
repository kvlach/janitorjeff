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
	h, err := helix.NewClient(&helix.Options{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	})
	if err != nil {
		return nil, err
	}

	if at, rt, err := dbGetUserTokens(userID); err == nil {
		if at == "" || rt == "" {
			h.SetAppAccessToken(appAccessToken.Get())
		} else {
			h.SetUserAccessToken(at)
			h.SetRefreshToken(rt)
			h.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
				err := dbUpdateUserTokens(userID, newAccessToken, newRefreshToken)
				if err != nil {
					log.Error().
						Err(err).
						Msg("POSTGRES: failed to save new user tokens")
				}
			})
		}
	} else {
		h.SetAppAccessToken(appAccessToken.Get())
	}

	return &Helix{h}, nil
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

func (h *Helix) newAppAccessToken() error {
	// can't refresh an app access token, just get a new one
	err := generateAppAccessToken()
	if err != nil {
		return err
	}
	h.c.SetAppAccessToken(appAccessToken.Get())
	return nil
}

func (h *Helix) refreshToken() error {
	// the wrapper library handles refreshing user access tokens
	if h.c.GetUserAccessToken() != "" {
		return nil
	}
	return h.newAppAccessToken()
}

func (h *Helix) GetFollower(broadcasterID, userID string) (helix.UserFollow, error) {
	resp, err := h.c.GetUsersFollows(&helix.UsersFollowsParams{
		FromID: userID,
		ToID:   broadcasterID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Follows))

	switch err {
	case nil:
		return resp.Data.Follows[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.UserFollow{}, err
		}
		return h.GetFollower(broadcasterID, userID)

	default:
		return helix.UserFollow{}, err
	}
}

func (h *Helix) GetStream(broadcasterID string) (helix.Stream, error) {
	resp, err := h.c.GetStreams(&helix.StreamsParams{
		UserIDs: []string{broadcasterID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Streams))

	switch err {
	case nil:
		return resp.Data.Streams[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.Stream{}, err
		}
		return h.GetStream(broadcasterID)

	default:
		return helix.Stream{}, err
	}
}

func (h *Helix) GetUser(userID string) (helix.User, error) {
	resp, err := h.c.GetUsers(&helix.UsersParams{
		IDs: []string{userID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Users))

	switch err {
	case nil:
		return resp.Data.Users[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.User{}, err
		}
		return h.GetUser(userID)

	default:
		return helix.User{}, err
	}
}

func (h *Helix) GetUserID(username string) (string, error) {
	resp, err := h.c.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Users))

	switch err {
	case nil:
		return resp.Data.Users[0].ID, nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return "", err
		}
		return h.GetUserID(username)

	default:
		return "", err
	}
}

func (h *Helix) GetClip(clipID string) (helix.Clip, error) {
	resp, err := h.c.GetClips(&helix.ClipsParams{
		IDs: []string{clipID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Clips))

	switch err {
	case nil:
		return resp.Data.Clips[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.Clip{}, err
		}
		return h.GetClip(clipID)

	default:
		return helix.Clip{}, err
	}
}

func (h *Helix) GetBannedUser(broadcasterID, userID string) (helix.Ban, error) {
	resp, err := h.c.GetBannedUsers(&helix.BannedUsersParams{
		BroadcasterID: broadcasterID,
		UserID:        userID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Bans))

	switch err {
	case nil:
		return resp.Data.Bans[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.Ban{}, err
		}
		return h.GetBannedUser(broadcasterID, userID)

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

func (h *Helix) SearchGame(gameName string) (helix.Game, error) {
	resp, err := h.c.GetGames(&helix.GamesParams{
		Names: []string{gameName},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Games))

	switch err {
	case nil:
		return resp.Data.Games[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.Game{}, err
		}
		return h.SearchGame(gameName)

	default:
		return helix.Game{}, err
	}
}

func (h *Helix) GetChannelInfo(broadcasterID string) (helix.ChannelInformation, error) {
	resp, err := h.c.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{broadcasterID},
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Channels))

	switch err {
	case nil:
		return resp.Data.Channels[0], nil

	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return helix.ChannelInformation{}, err
		}
		return h.GetChannelInfo(broadcasterID)

	default:
		return helix.ChannelInformation{}, err
	}
}

func (h *Helix) GetGameName(channelId string) (string, error) {
	ch, err := h.GetChannelInfo(channelId)
	return ch.GameName, err
}

func (h *Helix) GetTitle(channelId string) (string, error) {
	ch, err := h.GetChannelInfo(channelId)
	return ch.Title, err
}

func (h *Helix) EditChannelInfo(broadcasterID, title, gameID string) (error, error) {
	if h.c.GetUserAccessToken() == "" {
		return ErrUserTokenRequired, nil
	}

	// both the title and the game need to be set at the same time
	resp, err := h.c.EditChannelInformation(&helix.EditChannelInformationParams{
		BroadcasterID: broadcasterID,
		Title:         title,
		GameID:        gameID,
	})

	err = checkErrors(err, resp.ResponseCommon, 1)

	switch err {
	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return nil, err
		}
		return h.EditChannelInfo(broadcasterID, title, gameID)

	default:
		return nil, err
	}
}

func (h *Helix) SetTitle(channelID, title string) (error, error) {
	ch, err := h.GetChannelInfo(channelID)
	if err != nil {
		return nil, err
	}
	return h.EditChannelInfo(channelID, title, ch.GameID)
}

func (h *Helix) SetGame(channelID, gameName string) (string, error, error) {
	title, err := h.GetTitle(channelID)
	if err != nil {
		return "", nil, err
	}

	// Clears the game
	if gameName == "-" {
		urr, err := h.EditChannelInfo(channelID, title, "0")
		return "nothing", urr, err
	}

	g, err := h.SearchGame(gameName)
	if err != nil {
		return "", nil, err
	}
	urr, err := h.EditChannelInfo(channelID, title, g.ID)
	return g.Name, urr, err
}

func (h *Helix) RedeemsList(broadcasterID string) ([]helix.ChannelCustomReward, error) {
	resp, err := h.c.GetCustomRewards(&helix.GetCustomRewardsParams{
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
		if err := h.refreshToken(); err != nil {
			return nil, err
		}
		return h.RedeemsList(broadcasterID)
	default:
		return nil, err
	}
}

func (h *Helix) CreateSubscription(broadcasterID, t string) (string, error) {
	resp, err := h.c.CreateEventSubSubscription(&helix.EventSubSubscription{
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
		if err := h.refreshToken(); err != nil {
			return "", err
		}
		return h.CreateSubscription(broadcasterID, t)
	default:
		return "", err
	}
}

func (h *Helix) DeleteSubscription(subID string) error {
	resp, err := h.c.RemoveEventSubSubscription(subID)
	if err != nil {
		return err
	}

	err = checkErrors(err, resp.ResponseCommon, 1)

	switch err {
	case nil:
		return nil
	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return err
		}
		return h.DeleteSubscription(subID)
	default:
		return err
	}
}

func (h *Helix) ListSubscriptions() ([]helix.EventSubSubscription, error) {
	resp, err := h.c.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{})
	if err != nil {
		return nil, err
	}

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.EventSubSubscriptions))

	switch err {
	case nil:
		return resp.Data.EventSubSubscriptions, nil
	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return nil, err
		}
		return h.ListSubscriptions()
	default:
		return nil, err
	}
}

func (h *Helix) IsMod(broadcasterID, userID string) (bool, error) {
	resp, err := h.c.GetModerators(&helix.GetModeratorsParams{
		BroadcasterID: broadcasterID,
		UserIDs:       []string{userID},
	})

	switch err := checkErrors(err, resp.ResponseCommon, 1); err {
	case nil:
		return len(resp.Data.Moderators) > 1, nil
	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return false, err
		}
		return h.IsMod(broadcasterID, userID)
	default:
		return false, err
	}
}

func (h *Helix) IsSub(broadcasterID, userID string) (bool, error) {
	resp, err := h.c.CheckUserSubscription(&helix.UserSubscriptionsParams{
		BroadcasterID: broadcasterID,
		UserID:        userID,
	})

	switch err := checkErrors(err, resp.ResponseCommon, 1); err {
	case nil:
		return len(resp.Data.UserSubscriptions) > 1, nil
	case ErrRetry:
		if err := h.refreshToken(); err != nil {
			return false, err
		}
		return h.IsSub(broadcasterID, userID)
	default:
		return false, err
	}
}
