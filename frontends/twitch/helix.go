package twitch

import (
	"fmt"
	"time"

	"github.com/nicklaw5/helix"
)

type Helix struct {
	*helix.Client
}

func HelixInit(clientID, token string) (*Helix, error) {
	h, err := helix.NewClient(&helix.Options{
		ClientID:        clientID,
		UserAccessToken: token,
	})

	return &Helix{h}, err
}

func checkErrors(err error, resp helix.ResponseCommon, length int) error {
	if err != nil {
		return fmt.Errorf("helix error: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%d, %s: %s", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	if length == 0 {
		return fmt.Errorf("no results")
	}

	return nil
}

func (h *Helix) GetFollower(channel_id, user_id string) (helix.UserFollow, error) {
	resp, err := h.GetUsersFollows(&helix.UsersFollowsParams{
		FromID: user_id,
		ToID:   channel_id,
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Follows)); err != nil {
		return helix.UserFollow{}, fmt.Errorf("failed to get follow date in channel '%s' for user '%s': %v", channel_id, user_id, err)
	}

	return resp.Data.Follows[0], nil
}

func (h *Helix) GetStream(channel_id string) (helix.Stream, error) {
	resp, err := h.GetStreams(&helix.StreamsParams{
		UserIDs: []string{channel_id},
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Streams)); err != nil {
		return helix.Stream{}, fmt.Errorf("failed to get stream info for channel '%s': %v", channel_id, err)
	}

	return resp.Data.Streams[0], nil
}

func (h *Helix) GetStreamStartedDate(channel_id string) (time.Time, error) {
	s, err := h.GetStream(channel_id)
	return s.StartedAt, err
}

func (h *Helix) GetUser(user_id string) (helix.User, error) {
	resp, err := h.GetUsers(&helix.UsersParams{
		IDs: []string{user_id},
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Users)); err != nil {
		return helix.User{}, fmt.Errorf("failed to get user info for id '%s': %v", user_id, err)
	}

	return resp.Data.Users[0], nil
}

func (h *Helix) GetUserName(user_id string) (string, error) {
	u, err := h.GetUser(user_id)
	return u.Login, err
}

func (h *Helix) GetUserDisplayName(user_id string) (string, error) {
	u, err := h.GetUser(user_id)
	return u.DisplayName, err
}

func (h *Helix) GetUserID(username string) (string, error) {
	resp, err := h.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Users)); err != nil {
		return "", fmt.Errorf("failed to get id for username '%s': %v", username, err)
	}

	return resp.Data.Users[0].ID, nil
}

func (h *Helix) GetClip(clip_id string) (helix.Clip, error) {
	resp, err := h.GetClips(&helix.ClipsParams{
		IDs: []string{clip_id},
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Clips)); err != nil {
		return helix.Clip{}, fmt.Errorf("failed to get clip with id '%s': %v", clip_id, err)
	}

	return resp.Data.Clips[0], nil
}

func (h *Helix) GetBannedUser(channel_id, user_id string) (helix.Ban, error) {
	resp, err := h.GetBannedUsers(&helix.BannedUsersParams{
		BroadcasterID: channel_id,
		UserID:        user_id,
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Bans)); err != nil {
		return helix.Ban{}, fmt.Errorf("failed to get banned user '%s' in channel '%s': %v", user_id, channel_id, err)
	}

	return resp.Data.Bans[0], nil
}

func (h *Helix) IsUserBanned(channel_id, user_id string) (bool, error) {
	resp, err := h.GetBannedUsers(&helix.BannedUsersParams{
		BroadcasterID: channel_id,
		UserID:        user_id,
	})

	if err := checkErrors(err, resp.ResponseCommon, 1); err != nil {
		return false, fmt.Errorf("failed to get banned user '%s' in channel '%s': %v", user_id, channel_id, err)
	}

	if len(resp.Data.Bans) > 1 {
		return true, nil
	}
	return false, nil
}

func (h *Helix) SearchGame(game_name string) (helix.Game, error) {
	resp, err := h.GetGames(&helix.GamesParams{
		Names: []string{game_name},
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Games)); err != nil {
		return helix.Game{}, fmt.Errorf("failed to search game '%s': %v", game_name, err)
	}

	return resp.Data.Games[0], nil
}

func (h *Helix) GetChannelInfo(channel_id string) (helix.ChannelInformation, error) {
	resp, err := h.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterID: channel_id,
	})

	if err := checkErrors(err, resp.ResponseCommon, len(resp.Data.Channels)); err != nil {
		return helix.ChannelInformation{}, fmt.Errorf("failed to get inforamtion for channel '%s': %v", channel_id, err)
	}

	return resp.Data.Channels[0], nil
}

func (h *Helix) GetGameName(channel_id string) (string, error) {
	ch, err := h.GetChannelInfo(channel_id)
	return ch.GameName, err
}

func (h *Helix) GetTitle(channel_id string) (string, error) {
	ch, err := h.GetChannelInfo(channel_id)
	return ch.Title, err
}

func (h *Helix) EditChannelInfo(channel_id, title, game_id string) error {
	// both the title and the game need to be set at the same time
	resp, err := h.EditChannelInformation(&helix.EditChannelInformationParams{
		BroadcasterID: channel_id,
		Title:         title,
		GameID:        game_id,
	})

	if err := checkErrors(err, resp.ResponseCommon, 1); err != nil {
		return fmt.Errorf("failed to update inforamtion for channel '%s' to title '%s' and game '%s'", channel_id, title, game_id)
	}

	return nil
}

func (h *Helix) SetTitle(channel_id, title string) error {
	ch, err := h.GetChannelInfo(channel_id)
	if err != nil {
		return err
	}

	return h.EditChannelInfo(channel_id, title, ch.GameID)
}

func (h *Helix) SetGame(channel_id, game_name string) (string, error) {
	title, err := h.GetTitle(channel_id)
	if err != nil {
		return "", err
	}

	// Clears the game
	if game_name == "-" {
		return "nothing", h.EditChannelInfo(channel_id, title, "0")
	}

	g, err := h.SearchGame(game_name)
	if err != nil {
		return "", err
	}

	return g.Name, h.EditChannelInfo(channel_id, title, g.ID)
}
