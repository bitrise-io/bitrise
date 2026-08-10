package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestProfileService_Me(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("path = %q, want /me", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"username":"alice","email":"alice@example.com","avatar_url":"https://example.com/a.png"}}`))
	})
	client := mustClient(t, srv.URL)

	profile, err := NewProfileService(client).Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Profile{
		Username:  "alice",
		Email:     "alice@example.com",
		AvatarURL: "https://example.com/a.png",
	}, profile)
}

// TestProfile_YAMLKeysMatchJSON locks in the yaml tags: yaml.v2 ignores json
// tags, so without them `user me --format yml` would emit `avatarurl`.
func TestProfile_YAMLKeysMatchJSON(t *testing.T) {
	out, err := yaml.Marshal(Profile{Username: "alice", Email: "a@b.io", AvatarURL: "https://example.com/a.png"})
	require.NoError(t, err)
	assert.Contains(t, string(out), "avatar_url:")
	assert.NotContains(t, string(out), "avatarurl:")
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func mustClient(t *testing.T, baseURL string) *bitriseapi.Client {
	t.Helper()
	c, err := bitriseapi.New(baseURL, "t")
	require.NoError(t, err)
	return c
}
