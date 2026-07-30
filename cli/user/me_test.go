package user

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestMeCmd_HappyPath(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":{"username":"alice","email":"alice@example.com"}}`))
	}))
	t.Cleanup(srv.Close)

	cmd := newTestMeCmd(t, srv.URL)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/me", gotPath)
	assert.Equal(t, "token test-token", gotAuth)
}

func TestMeCmd_NoTokenErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewMeCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetContext(t.Context())

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
}

func newTestMeCmd(t *testing.T, apiBaseURL string) *cobra.Command {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewMeCommand()
	cmd.SetOut(&bytes.Buffer{})
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd
}
