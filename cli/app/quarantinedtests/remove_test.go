package quarantinedtests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
)

func TestRemove_WithYes(t *testing.T) {
	var requests []string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})

	out, _, err := runCmd(t, NewRemoveCommand(), srv.URL, "qt-1", "--yes")
	require.NoError(t, err)
	assert.Equal(t, []string{"DELETE /apps/my-app/quarantined-tests/qt-1"}, requests, "--yes skips the lookup for the prompt")
	assert.Equal(t, "Removed quarantined test qt-1\n", out)
}

func TestRemove_Prompt(t *testing.T) {
	tests := []struct {
		name       string
		answer     string
		wantRemove bool
	}{
		{name: "y", answer: "y\n", wantRemove: true},
		{name: "yes", answer: "yes\n", wantRemove: true},
		{name: "n", answer: "n\n"},
		{name: "empty", answer: "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTerminal(t)
			var deleted bool
			srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					deleted = true
					w.WriteHeader(http.StatusNoContent)
					return
				}
				_, _ = w.Write([]byte(quarantinedTestJSON))
			})

			cmd := NewRemoveCommand()
			cmd.SetIn(strings.NewReader(tt.answer))
			out, errOut, err := runCmd(t, cmd, srv.URL, "qt-1")

			assert.Contains(t, errOut, "Remove AppTests › LoginTests › LoginSpec › test_login from quarantine? [y/N]: ")
			assert.Equal(t, tt.wantRemove, deleted)
			if tt.wantRemove {
				require.NoError(t, err)
				assert.Equal(t, "Removed quarantined test qt-1\n", out)
			} else {
				assert.EqualError(t, err, "aborted")
				assert.Empty(t, out)
			}
		})
	}
}

func TestRemove_PromptLeavesOutClassNameTheNameStartsWith(t *testing.T) {
	fakeTerminal(t)
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"qt-2","test_case_name":"Settings shows a toast","class_name":"Settings shows a toast","test_suite_name":["Settings"]}`))
	})

	cmd := NewRemoveCommand()
	cmd.SetIn(strings.NewReader("n\n"))
	_, errOut, _ := runCmd(t, cmd, srv.URL, "qt-2")
	assert.Contains(t, errOut, "Remove Settings › Settings shows a toast from quarantine? [y/N]: ")
}

func TestRemove_NonTerminalNeedsYes(t *testing.T) {
	srv := newFakeServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("no request expected")
	})

	cmd := NewRemoveCommand()
	cmd.SetIn(strings.NewReader("y\n"))
	_, _, err := runCmd(t, cmd, srv.URL, "qt-1")
	assert.ErrorContains(t, err, "pass --yes")
}

func TestRemove_NotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error_msg":"Not Found"}`))
	})

	_, _, err := runCmd(t, NewRemoveCommand(), srv.URL, "missing", "--yes")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "removing quarantined test failed")
}

// fakeTerminal makes stdin count as a terminal, so remove prompts on the
// reader the test passes to SetIn.
func fakeTerminal(t *testing.T) {
	t.Helper()
	orig := cmdutil.IsTerminal
	cmdutil.IsTerminal = func(io.Reader) bool { return true }
	t.Cleanup(func() { cmdutil.IsTerminal = orig })
}
