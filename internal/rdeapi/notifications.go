package rdeapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// SessionNotification is a persisted notification from a session VM
// (agent stop/idle/permission prompts, etc.).
type SessionNotification struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId,omitempty"`
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	Type      string `json:"type,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// ListSessionNotificationsOptions filters the notifications list. All
// fields are optional. Server defaults: limit=50, order=desc.
type ListSessionNotificationsOptions struct {
	CreatedBefore string // RFC3339; only notifications created strictly before this
	CreatedAfter  string // RFC3339; only notifications created strictly after this
	Limit         int
	Order         string // "asc" or "desc"; empty means server default (desc)
}

// ListSessionNotifications returns notifications for a session.
// Endpoint: GET /v1/workspaces/{workspaceId}/sessions/{sessionId}/notifications.
func (c *Client) ListSessionNotifications(ctx context.Context, workspaceID, sessionID string, opts ListSessionNotificationsOptions) ([]SessionNotification, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/notifications")
	q := url.Values{}
	if opts.CreatedBefore != "" {
		q.Set("createdBefore", opts.CreatedBefore)
	}
	if opts.CreatedAfter != "" {
		q.Set("createdAfter", opts.CreatedAfter)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	switch opts.Order {
	case "asc":
		q.Set("order", "SORT_ORDER_ASC")
	case "desc":
		q.Set("order", "SORT_ORDER_DESC")
	}
	if encoded := q.Encode(); encoded != "" {
		p += "?" + encoded
	}

	var resp listSessionNotificationsResp
	if err := c.getJSON(ctx, p, &resp); err != nil {
		return nil, err
	}
	return resp.Notifications, nil
}

type listSessionNotificationsResp struct {
	Notifications []SessionNotification `json:"notifications"`
}
