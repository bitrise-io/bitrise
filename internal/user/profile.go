package user

import (
	"context"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// ProfileService looks up the authenticated user's profile via the Bitrise
// API, unlike Service, which talks to app.bitrise.io directly.
type ProfileService struct {
	client *bitriseapi.Client
}

func NewProfileService(client *bitriseapi.Client) *ProfileService {
	return &ProfileService{client: client}
}

// Me returns the profile of the authenticated user.
func (s *ProfileService) Me(ctx context.Context) (bitriseapi.User, error) {
	return s.client.Me(ctx)
}
