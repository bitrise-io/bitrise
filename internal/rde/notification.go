package rde

import (
	"context"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

type SessionNotification struct {
	ID        string     `json:"id" yaml:"id"`
	SessionID string     `json:"session_id,omitempty" yaml:"session_id,omitempty"`
	Title     string     `json:"title,omitempty" yaml:"title,omitempty"`
	Body      string     `json:"body,omitempty" yaml:"body,omitempty"`
	Type      string     `json:"type,omitempty" yaml:"type,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

// ListSessionNotificationsOptions paginates and filters the notifications list.
type ListSessionNotificationsOptions struct {
	CreatedBefore string // RFC3339
	CreatedAfter  string // RFC3339
	Limit         int
	Order         string // "asc" | "desc" | ""
}

func (s *Service) ListSessionNotifications(ctx context.Context, workspaceID, sessionID string, opts ListSessionNotificationsOptions) ([]SessionNotification, error) {
	if s.client == nil {
		return nil, errClient()
	}
	wire, err := s.client.ListSessionNotifications(ctx, workspaceID, sessionID, rdeapi.ListSessionNotificationsOptions{
		CreatedBefore: opts.CreatedBefore,
		CreatedAfter:  opts.CreatedAfter,
		Limit:         opts.Limit,
		Order:         opts.Order,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SessionNotification, 0, len(wire))
	for _, w := range wire {
		out = append(out, SessionNotification{
			ID:        w.ID,
			SessionID: w.SessionID,
			Title:     w.Title,
			Body:      w.Body,
			Type:      notificationTypeFromAPI(w.Type),
			CreatedAt: parseTime(w.CreatedAt),
		})
	}
	return out, nil
}

func notificationTypeFromAPI(s string) string {
	if s == "" {
		return ""
	}
	const prefix = "SESSION_NOTIFICATION_TYPE_"
	v := strings.TrimPrefix(s, prefix)
	if v == "UNSPECIFIED" {
		return ""
	}
	return strings.ToLower(v)
}
