package twitch

import (
	"errors"
	"fmt"

	"github.com/nicklaw5/helix"
	"github.com/rs/zerolog/log"
)

var (
	ErrExpiredRefreshToken = errors.New("The user will need to reconnect the bot to twitch.")
	ErrRetry               = errors.New("Refresh the access token and try again.")
	ErrNoResults           = errors.New("Couldn't find what you were looking for.")
)

func RefreshUserAccessToken(accessToken string) (string, error) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	})
	if err != nil {
		return "", err
	}

	refreshToken, err := dbGetetUserRefreshToken(accessToken)
	if err != nil {
		return "", err
	}

	resp, err := client.RefreshUserAccessToken(refreshToken)
	log.Debug().
		Err(err).
		Str("old", accessToken).
		Msg("refreshed user access token")
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 401 {
		return "", ErrExpiredRefreshToken
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%d, %s: %s", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	err = dbUpdateUserTokens(accessToken, resp.Data.AccessToken, resp.Data.RefreshToken)
	return resp.Data.AccessToken, err
}

type Helix struct {
	c *helix.Client
}

func HelixInit(token string) (*Helix, error) {
	h, err := helix.NewClient(&helix.Options{
		ClientID:        ClientID,
		UserAccessToken: token,
	})

	return &Helix{h}, err
}

func checkErrors(err error, resp helix.ResponseCommon, length int) error {
	if err != nil {
		return fmt.Errorf("helix error: %v", err)
	}

	if resp.StatusCode == 401 {
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

func (h *Helix) refreshUserAccessToken() error {
	token, err := RefreshUserAccessToken(h.c.GetUserAccessToken())
	if err != nil {
		return err
	}
	h.c.SetUserAccessToken(token)
	return nil
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
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
		if err := h.refreshUserAccessToken(); err != nil {
			return helix.Game{}, err
		}
		return h.SearchGame(gameName)

	default:
		return helix.Game{}, err
	}
}

func (h *Helix) GetChannelInfo(broadcasterID string) (helix.ChannelInformation, error) {
	resp, err := h.c.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterID: broadcasterID,
	})

	err = checkErrors(err, resp.ResponseCommon, len(resp.Data.Channels))

	switch err {
	case nil:
		return resp.Data.Channels[0], nil

	case ErrRetry:
		if err := h.refreshUserAccessToken(); err != nil {
			return helix.ChannelInformation{}, err
		}
		return h.GetChannelInfo(broadcasterID)

	default:
		return helix.ChannelInformation{}, err
	}
}

func (h *Helix) GetGameName(channel_id string) (string, error) {
	ch, err := h.GetChannelInfo(channel_id)
	return ch.GameName, err
}

func (h *Helix) GetTitle(channel_id string) (string, error) {
	ch, err := h.GetChannelInfo(channel_id)
	return ch.Title, err
}

func (h *Helix) EditChannelInfo(broadcasterID, title, gameID string) error {
	// both the title and the game need to be set at the same time
	resp, err := h.c.EditChannelInformation(&helix.EditChannelInformationParams{
		BroadcasterID: broadcasterID,
		Title:         title,
		GameID:        gameID,
	})

	err = checkErrors(err, resp.ResponseCommon, 1)

	switch err {
	case ErrRetry:
		if err := h.refreshUserAccessToken(); err != nil {
			return err
		}
		return h.EditChannelInfo(broadcasterID, title, gameID)

	default:
		return err
	}
}

func (h *Helix) SetTitle(channelID, title string) error {
	ch, err := h.GetChannelInfo(channelID)
	if err != nil {
		return err
	}
	return h.EditChannelInfo(channelID, title, ch.GameID)
}

func (h *Helix) SetGame(channelID, gameName string) (string, error) {
	title, err := h.GetTitle(channelID)
	if err != nil {
		return "", err
	}

	// Clears the game
	if gameName == "-" {
		return "nothing", h.EditChannelInfo(channelID, title, "0")
	}

	g, err := h.SearchGame(gameName)
	if err != nil {
		return "", err
	}

	return g.Name, h.EditChannelInfo(channelID, title, g.ID)
}
