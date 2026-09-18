package session

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestBuildExecCommand(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		shell bool
		want  string
	}{
		{
			name: "default quotes each token so metacharacters stay literal",
			args: []string{"echo", "a; b"},
			want: `echo 'a; b'`,
		},
		{
			name: "default leaves simple argv untouched",
			args: []string{"npm", "test"},
			want: "npm test",
		},
		{
			name:  "shell mode joins a single quoted command verbatim",
			args:  []string{"cd repo && ls | head"},
			shell: true,
			want:  "cd repo && ls | head",
		},
		{
			name:  "shell mode joins multiple tokens with spaces, no quoting",
			args:  []string{"cd", "/x", "&&", "ls"},
			shell: true,
			want:  "cd /x && ls",
		},
		{
			// Shell mode does not escape single quotes — that is the remote
			// wrapper's job (see internalrde.buildLoginShellCmd). Here the
			// command is passed through verbatim, quotes and all.
			name:  "shell mode passes single quotes through unescaped",
			args:  []string{"grep -rn 'TODO' ."},
			shell: true,
			want:  "grep -rn 'TODO' .",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildExecCommand(tc.args, tc.shell); got != tc.want {
				t.Errorf("buildExecCommand(%q, shell=%v) = %q, want %q", tc.args, tc.shell, got, tc.want)
			}
		})
	}
}

// TestExecCmd_TimeoutFlagDefault pins that exec registers --timeout and that
// its default tracks the service-layer default (so a change to one without the
// other is caught).
func TestExecCmd_TimeoutFlagDefault(t *testing.T) {
	c := newExecCmd()
	if c.Flags().Lookup("timeout") == nil {
		t.Fatal("exec should register a --timeout flag")
	}
	got, err := c.Flags().GetDuration("timeout")
	if err != nil {
		t.Fatalf("GetDuration(timeout): %v", err)
	}
	if got != internalrde.DefaultExecuteTimeout {
		t.Errorf("--timeout default = %s, want %s", got, internalrde.DefaultExecuteTimeout)
	}
	if internalrde.DefaultExecuteTimeout != 10*time.Minute {
		t.Errorf("DefaultExecuteTimeout = %s, want 10m", internalrde.DefaultExecuteTimeout)
	}
}

func TestExecCmd_EnvFlagsRegistered(t *testing.T) {
	c := newExecCmd()
	envFlag := c.Flags().Lookup("env")
	if envFlag == nil {
		t.Fatal("exec should register an --env flag")
	}
	if envFlag.Value.Type() != "stringArray" {
		t.Errorf("--env type = %q, want stringArray (repeatable, no comma-splitting)", envFlag.Value.Type())
	}
	noFile, err := c.Flags().GetBool("no-env-file")
	if err != nil {
		t.Fatalf("GetBool(no-env-file): %v", err)
	}
	if noFile {
		t.Error("--no-env-file should default to false")
	}
}

// TestExecCmd_EnvFlagUnsetVarFailsFast pins that a --env NAME with no local
// value errors before any network call — the server fails the test on any
// request — and that the error names the variable.
func TestExecCmd_EnvFlagUnsetVarFailsFast(t *testing.T) {
	t.Chdir(t.TempDir()) // no dotfile in scope
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request before env validation: %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)

	_, _, err := cmdtest.Run(t, newExecCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--env", "BCLI_TEST_DEFINITELY_UNSET", "--", "echo", "hi"},
	})
	if err == nil || !strings.Contains(err.Error(), "BCLI_TEST_DEFINITELY_UNSET") {
		t.Fatalf("err = %v, want error naming BCLI_TEST_DEFINITELY_UNSET", err)
	}
}

func TestExecCmd_DotfileWarningAndNotice(t *testing.T) {
	writeExecEnvDotfile(t, "exec:\n  env:\n    - BCLI_TEST_DEFINITELY_UNSET\n    - FOO=secret-value-xyz\n")
	srv := terminatedSessionServer(t)

	_, stderr, err := cmdtest.Run(t, newExecCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--", "echo", "hi"},
	})
	if err == nil {
		t.Fatal("expected the exec against a terminated session to error")
	}
	if !strings.Contains(stderr, "BCLI_TEST_DEFINITELY_UNSET not set locally") {
		t.Errorf("stderr missing skip warning:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Forwarding env: FOO") {
		t.Errorf("stderr missing forwarding notice:\n%s", stderr)
	}
	if strings.Contains(stderr, "secret-value-xyz") {
		t.Errorf("stderr leaks a forwarded value:\n%s", stderr)
	}
}

func TestExecCmd_NoEnvFileFlag(t *testing.T) {
	writeExecEnvDotfile(t, "exec:\n  env:\n    - BCLI_TEST_DEFINITELY_UNSET\n    - FOO=secret-value-xyz\n")
	srv := terminatedSessionServer(t)

	_, stderr, err := cmdtest.Run(t, newExecCmd(), cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "--no-env-file", "--", "echo", "hi"},
	})
	if err == nil {
		t.Fatal("expected the exec against a terminated session to error")
	}
	for _, unwanted := range []string{"not set locally", "Forwarding env"} {
		if strings.Contains(stderr, unwanted) {
			t.Errorf("stderr should not mention %q with --no-env-file:\n%s", unwanted, stderr)
		}
	}
}

// TestExecCmd_QuietSuppressesNoticeNotWarning attaches the persistent --quiet
// flag directly to the command under test (cmdutil.IsQuiet reads it off
// cmd.Root(), which for a standalone command is the command itself): with -q
// the forwarding notice disappears while the skip warning stays, per the
// output scheme's rule that warnings ignore --quiet.
func TestExecCmd_QuietSuppressesNoticeNotWarning(t *testing.T) {
	writeExecEnvDotfile(t, "exec:\n  env:\n    - BCLI_TEST_DEFINITELY_UNSET\n    - FOO=secret-value-xyz\n")
	srv := terminatedSessionServer(t)

	cmd := newExecCmd()
	cmd.PersistentFlags().BoolP(cmdutil.FlagQuiet, "q", false, "quiet")

	_, stderr, err := cmdtest.Run(t, cmd, cmdtest.Opts{
		RDEAPIBaseURL:      srv.URL,
		DefaultWorkspaceID: "ws-1",
		Args:               []string{uuidSession, "-q", "--", "echo", "hi"},
	})
	if err == nil {
		t.Fatal("expected the exec against a terminated session to error")
	}
	if strings.Contains(stderr, "Forwarding env") {
		t.Errorf("-q should suppress the forwarding notice:\n%s", stderr)
	}
	if !strings.Contains(stderr, "BCLI_TEST_DEFINITELY_UNSET not set locally") {
		t.Errorf("-q must not suppress the skip warning:\n%s", stderr)
	}
}

// TestRenderExecResult_NonRawFormats pins the format-predicate translation
// (format != FormatRaw, not == FormatJSON as in the reference implementation,
// which never had a third format to consider): both json and yml render the
// full structured result rather than falling through to the raw stdout/stderr
// streaming branch.
func TestRenderExecResult_NonRawFormats(t *testing.T) {
	res := internalrde.ExecResult{ExitCode: 0, Stdout: "hi\n", Stderr: ""}
	for _, format := range []string{output.FormatJSON, output.FormatYML} {
		t.Run(format, func(t *testing.T) {
			old := output.Format
			output.Format = format
			defer func() { output.Format = old }()

			cmd := &cobra.Command{}
			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := renderExecResult(cmd, res); err != nil {
				t.Fatalf("renderExecResult: %v", err)
			}
			if !strings.Contains(out.String(), "hi") {
				t.Errorf("%s output missing stdout field:\n%s", format, out.String())
			}
		})
	}
}

// writeExecEnvDotfile drops a .bitrise/rde.yml into dir and chdirs there, so
// the test exercises the auto-read path without picking up any real dotfile
// in an ancestor of the package directory.
func writeExecEnvDotfile(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".bitrise"), 0o755); err != nil { // test-only tempdir, perms do not matter
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".bitrise", "rde.yml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

// terminatedSessionServer serves a GET for uuidSession whose status is
// terminated, so exec's Execute call fails cleanly after the env diagnostics
// have been printed (no SSH dial is attempted against a terminated session).
func terminatedSessionServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/workspaces/ws-1/sessions/"+uuidSession {
			_, _ = io.WriteString(w, `{"session":{"id":"s-1","name":"dev","status":"SESSION_STATUS_TERMINATED"}}`)
			return
		}
		t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	return srv
}
