package rdeapi

import (
	"context"
	"net/http"
	"testing"
)

func TestListSessionNotifications_PathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"notifications":[
		{"id":"n1","sessionId":"s1","title":"Agent idle","type":"IDLE","createdAt":"2026-05-28T00:00:00Z"}
	]}`)

	notifs, err := rs.client().ListSessionNotifications(context.Background(), "ws-1", "s1", ListSessionNotificationsOptions{})
	if err != nil {
		t.Fatalf("ListSessionNotifications: %v", err)
	}
	if rs.lastMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/notifications"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(notifs) != 1 || notifs[0].Title != "Agent idle" {
		t.Errorf("notifications = %+v", notifs)
	}
}

func TestListSessionNotifications_NoOptionsMeansNoQuery(t *testing.T) {
	rs := newRecordingServer(t, `{"notifications":[]}`)

	if _, err := rs.client().ListSessionNotifications(context.Background(), "ws-1", "s1", ListSessionNotificationsOptions{}); err != nil {
		t.Fatalf("ListSessionNotifications: %v", err)
	}
	if rs.lastQuery != "" {
		t.Errorf("expected no query params, got %q", rs.lastQuery)
	}
}

func TestListSessionNotifications_EncodesAllOptions(t *testing.T) {
	rs := newRecordingServer(t, `{"notifications":[]}`)

	_, err := rs.client().ListSessionNotifications(context.Background(), "ws-1", "s1", ListSessionNotificationsOptions{
		CreatedBefore: "2026-05-28T12:00:00Z",
		CreatedAfter:  "2026-05-01T00:00:00Z",
		Limit:         25,
		Order:         "asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	q := parseQuery(t, rs.lastQuery)
	checks := map[string]string{
		"createdBefore": "2026-05-28T12:00:00Z",
		"createdAfter":  "2026-05-01T00:00:00Z",
		"limit":         "25",
		"order":         "SORT_ORDER_ASC", // friendly "asc" maps to the API enum
	}
	for k, want := range checks {
		if got := q.Get(k); got != want {
			t.Errorf("query %q = %q, want %q", k, got, want)
		}
	}
}

func TestListSessionNotifications_OrderEnumMapping(t *testing.T) {
	cases := map[string]string{
		"asc":     "SORT_ORDER_ASC",
		"desc":    "SORT_ORDER_DESC",
		"":        "", // unset → server default, no order param
		"unknown": "", // unrecognized → dropped, not passed through raw
	}
	for in, want := range cases {
		t.Run("order="+in, func(t *testing.T) {
			rs := newRecordingServer(t, `{"notifications":[]}`)
			if _, err := rs.client().ListSessionNotifications(context.Background(), "ws-1", "s1", ListSessionNotificationsOptions{Order: in}); err != nil {
				t.Fatal(err)
			}
			q := parseQuery(t, rs.lastQuery)
			if got := q.Get("order"); got != want {
				t.Errorf("order=%q -> %q, want %q", in, got, want)
			}
		})
	}
}

func TestListSessionNotifications_OmitsZeroLimit(t *testing.T) {
	rs := newRecordingServer(t, `{"notifications":[]}`)

	if _, err := rs.client().ListSessionNotifications(context.Background(), "ws-1", "s1", ListSessionNotificationsOptions{Limit: 0}); err != nil {
		t.Fatal(err)
	}
	q := parseQuery(t, rs.lastQuery)
	if _, ok := q["limit"]; ok {
		t.Errorf("limit=0 should be omitted, query = %q", rs.lastQuery)
	}
}

func TestNotifications_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	if _, err := c.ListSessionNotifications(ctx, "", "s1", ListSessionNotificationsOptions{}); err == nil {
		t.Error("expected error for empty workspace ID")
	}
	if _, err := c.ListSessionNotifications(ctx, "ws", "", ListSessionNotificationsOptions{}); err == nil {
		t.Error("expected error for empty session ID")
	}
	if rs.hits != 0 {
		t.Errorf("validation guards made %d HTTP call(s); should short-circuit", rs.hits)
	}
}
