package yml

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

const validConfig = `format_version: "17"
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"
`

// locallyInvalidConfig fails local validation (missing format_version) but
// is still valid YAML, so it can be submitted online regardless — used to
// prove which path (local vs. online) actually produced a given result.
const locallyInvalidConfig = `workflows: {}
`

func TestValidateConfig_NoToken_UsesLocal(t *testing.T) {
	var onlineCalled bool
	srv := newValidateFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		onlineCalled = true
		_, _ = w.Write([]byte(`{"errors":[],"warnings":[]}`))
	})
	cmd := &cobra.Command{}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no auth.yaml written -> no token
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: srv.URL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))

	item, warning, err := validateConfig(cmd, "", encode(validConfig), false, "")
	require.NoError(t, err)
	assert.True(t, item.IsValid)
	assert.Empty(t, warning)
	assert.False(t, onlineCalled, "online validation must be skipped without a token")
}

func TestValidateConfig_Offline_UsesLocalEvenWithToken(t *testing.T) {
	var onlineCalled bool
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, _ *http.Request) {
		onlineCalled = true
		_, _ = w.Write([]byte(`{"errors":[],"warnings":[]}`))
	})

	item, warning, err := validateConfig(cmd, "", encode(validConfig), true, "")
	require.NoError(t, err)
	assert.True(t, item.IsValid)
	assert.Empty(t, warning)
	assert.False(t, onlineCalled, "--offline must skip online validation regardless of token presence")
}

func TestValidateConfig_NoConfigGiven_IsANoop(t *testing.T) {
	// No -c/--config-base64, and this package's directory has no bitrise.yml
	// for the default-path lookup to find either.
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("online validation must not be attempted when no config was given")
	})

	item, warning, err := validateConfig(cmd, "", "", false, "")
	require.NoError(t, err)
	assert.Nil(t, item)
	assert.Empty(t, warning)
}

func TestValidateConfig_OnlineSucceeds_SkipsLocalEntirely(t *testing.T) {
	var gotQuery string
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"errors":[],"warnings":["deprecated step used"]}`))
	})

	// Content is locally invalid (missing format_version); the fake server
	// reports it valid regardless. A valid result here can only have come
	// from online — proving local never ran alongside it.
	item, warning, err := validateConfig(cmd, "", encode(locallyInvalidConfig), false, "app-slug")
	require.NoError(t, err)
	assert.True(t, item.IsValid)
	assert.Equal(t, []string{"deprecated step used"}, item.Warnings)
	assert.Empty(t, warning, "no top-level warning expected when the online call itself succeeds")
	assert.Equal(t, "app_slug=app-slug", gotQuery)
}

func TestValidateConfig_Online422_UsesOnlyOnlineResult(t *testing.T) {
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"rejected by server-side schema"}`))
	})

	// Content is locally invalid for a different reason ("missing
	// format_version"); asserting the server's message (not local's) proves
	// this result came from online, not a local fallback.
	item, warning, err := validateConfig(cmd, "", encode(locallyInvalidConfig), false, "")
	require.NoError(t, err)
	assert.False(t, item.IsValid)
	assert.Equal(t, "rejected by server-side schema", item.Error)
	assert.Empty(t, warning, "a 422 is a completed online result, not a fallback case")
}

func TestValidateConfig_OnlineTransientFailure_FallsBackToLocal(t *testing.T) {
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream exploded"}`))
	})

	item, warning, err := validateConfig(cmd, "", encode(validConfig), false, "")
	require.NoError(t, err)
	assert.True(t, item.IsValid, "local validation must actually run as the fallback")
	assert.Contains(t, warning, "online validation unavailable")
}

func TestRunValidate_PropagatesConfigWarningIntoTopLevelWarnings(t *testing.T) {
	cmd := newTestValidateCmd(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream exploded"}`))
	})

	validation, warnings, err := runValidate(cmd, "", encode(validConfig), "", "", false, "")
	require.NoError(t, err)
	require.NotNil(t, validation.Config)
	assert.True(t, validation.Config.IsValid)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "online validation unavailable")
}

func encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func newValidateFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// newTestValidateCmd builds a bare *cobra.Command with a real token
// available (via a temp XDG_CONFIG_HOME + auth.yaml) and its context wired
// to handler's own httptest.Server.
func newTestValidateCmd(t *testing.T, handler http.HandlerFunc) *cobra.Command {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	apiBaseURL := newValidateFakeServer(t, handler).URL

	cmd := &cobra.Command{}
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd
}
