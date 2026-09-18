package bitriseapi

import (
	"context"
	"encoding/json"
	"fmt"
)

// User holds profile information for the authenticated user.
type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Me returns the profile of the authenticated user.
// Endpoint: GET /me. Unlike /available-stacks or the /bitrise.yml endpoints,
// the response is wrapped in a {"data": ...} envelope.
func (c *Client) Me(ctx context.Context) (User, error) {
	req, err := c.newRequest(ctx, "/me", nil)
	if err != nil {
		return User{}, err
	}

	body, err := c.do(req)
	if err != nil {
		return User{}, err
	}

	var envelope struct {
		Data User `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return User{}, fmt.Errorf("decode me response: %w", err)
	}
	return envelope.Data, nil
}
