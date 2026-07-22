package cmdutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCmd(t *testing.T, apiBaseURL string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd
}

func TestNewAPIClient_EnvTokenTakesPrecedenceOverAuthFile(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "file-token"}))
	t.Setenv("BITRISE_TOKEN", "env-token")

	client, err := NewAPIClient(newTestCmd(t, srv.URL))
	require.NoError(t, err)

	_, err = client.SearchSteps(context.Background(), bitriseapi.StepSearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, "token env-token", gotAuth)
}

func TestNewAPIClient_FallsBackToAuthFile(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "file-token"}))

	client, err := NewAPIClient(newTestCmd(t, srv.URL))
	require.NoError(t, err)

	_, err = client.SearchSteps(context.Background(), bitriseapi.StepSearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, "token file-token", gotAuth)
}

func TestNewAPIClient_ErrNoToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := NewAPIClient(newTestCmd(t, "https://api.example.test"))
	assert.ErrorIs(t, err, ErrNoToken)
}
