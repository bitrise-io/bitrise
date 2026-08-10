package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestCmd_GetRawBodyPassthrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/me", r.URL.Path)
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"data":{"username":"marton"}}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.RunE(cmd, []string{"/me"}))
	assert.Equal(t, `{"data":{"username":"marton"}}`+"\n", out.String())
}

func TestCmd_IncludePrintsStatusAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "yes")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("include", "true"))
	require.NoError(t, cmd.RunE(cmd, []string{"/me"}))

	got := out.String()
	assert.Contains(t, got, "HTTP 200 OK")
	assert.Contains(t, got, "X-Test: yes")
	assert.Contains(t, got, "{}")
}

func TestCmd_NonSuccessPrintsBodyAndErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestCmd(t, srv.URL)
	err := cmd.RunE(cmd, []string{"/missing"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, out.String(), "not found")
}

func TestCmd_MethodDefaultsToPostWhenFieldsPresent(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("field", "title=My App"))
	require.NoError(t, cmd.RunE(cmd, []string{"/apps"}))
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestCmd_ExplicitMethodIsUpcased(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("method", "delete"))
	require.NoError(t, cmd.RunE(cmd, []string{"/apps/APP_ID/builds/BUILD_ID"}))
	assert.Equal(t, http.MethodDelete, gotMethod)
}

func TestCmd_InputFromFile(t *testing.T) {
	var gotBody, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	path := filepath.Join(t.TempDir(), "body.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"hook_info":{"type":"bitrise"}}`), 0600))

	cmd, _ := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("input", path))
	require.NoError(t, cmd.RunE(cmd, []string{"/apps/APP_ID/builds"}))

	assert.Equal(t, `{"hook_info":{"type":"bitrise"}}`, gotBody)
	assert.Equal(t, "application/json", gotContentType)
}

func TestCmd_InputFromStdin(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestCmd(t, srv.URL)
	cmd.SetIn(strings.NewReader(`{"from":"stdin"}`))
	require.NoError(t, cmd.Flags().Set("input", "-"))
	require.NoError(t, cmd.RunE(cmd, []string{"/apps"}))
	assert.Equal(t, `{"from":"stdin"}`, gotBody)
}

func TestCmd_AllMergesPages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next") == "cursor2" {
			_, _ = w.Write([]byte(`{"data":[2],"paging":{"next":""}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[1],"paging":{"next":"cursor2"}}`))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("all", "true"))
	require.NoError(t, cmd.RunE(cmd, []string{"/apps"}))

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, []any{float64(1), float64(2)}, got["data"])
}

func TestCmd_FieldAndInputMutuallyExclusive(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"/apps", "--field", "a=b", "--input", "-"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field")
	assert.Contains(t, err.Error(), "input")
}

func TestCmd_RejectsMalformedFieldsAndHeaders(t *testing.T) {
	cases := map[string]struct{ flag, value string }{
		"field without separator": {flag: "field", value: "title"},
		"field with empty key":    {flag: "field", value: "=value"},
		"header without colon":    {flag: "header", value: "X-Test"},
		"header with empty name":  {flag: "header", value: " : value"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, _ := newTestCmd(t, "http://unused.invalid")
			require.NoError(t, cmd.Flags().Set(tc.flag, tc.value))

			err := cmd.RunE(cmd, []string{"/apps"})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "--"+tc.flag)
		})
	}
}

func TestCmd_RequiresPathArg(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "PATH argument should be required")
}

// newTestCmd builds a NewCmd() wired to apiBaseURL, with its output captured
// in the returned buffer.
func newTestCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// An ambient BITRISE_TOKEN outranks the saved one in cmdutil.ResolveToken.
	t.Setenv(auth.EnvToken, "")
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
