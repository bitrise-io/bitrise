package bitriseapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMe(t *testing.T) {
	var gotPath, gotAuth string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":{"username":"alice","email":"alice@example.com","avatar_url":"https://example.com/a.png"}}`))
	})

	user, err := newAPIClient(t, srv.URL, "my-token").Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/me", gotPath)
	assert.Equal(t, "token my-token", gotAuth)
	assert.Equal(t, User{Username: "alice", Email: "alice@example.com", AvatarURL: "https://example.com/a.png"}, user)
}

func TestMe_APIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	})

	_, err := newAPIClient(t, srv.URL, "bad-token").Me(context.Background())
	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
