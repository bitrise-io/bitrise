package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestProfileService_Me(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("path = %q, want /me", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"username":"alice","email":"alice@example.com"}}`))
	})
	client := bitriseapi.New(srv.URL, "t")

	profile, err := NewProfileService(client).Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, bitriseapi.User{Username: "alice", Email: "alice@example.com"}, profile)
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
