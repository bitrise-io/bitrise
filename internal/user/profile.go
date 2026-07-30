package user

import (
	"context"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// Profile is the CLI representation of the authenticated user.
type Profile struct {
	Username  string `json:"username" yaml:"username"`
	Email     string `json:"email" yaml:"email"`
	AvatarURL string `json:"avatar_url,omitempty" yaml:"avatar_url,omitempty"`
}

// ProfileService looks up the authenticated user's profile via the Bitrise
// API, unlike Service, which talks to app.bitrise.io directly.
type ProfileService struct {
	client *bitriseapi.Client
}

func NewProfileService(client *bitriseapi.Client) *ProfileService {
	return &ProfileService{client: client}
}

// Me returns the profile of the authenticated user.
func (s *ProfileService) Me(ctx context.Context) (Profile, error) {
	u, err := s.client.Me(ctx)
	if err != nil {
		return Profile{}, err
	}
	return Profile{
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
	}, nil
}
