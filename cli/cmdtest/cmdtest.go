// Package cmdtest is the shared test harness for rde command groups. It
// replaces five near-duplicate per-package `run` helpers (one apiece for
// session/template/savedinput/stack/machinetype in the reference
// implementation) with a single Run, adapted to how this repo actually
// resolves token/format/workspace: token from auth.yaml/BITRISE_TOKEN via
// cmdutil.ResolveToken, format from a real --format flag, workspace from
// flag/env/config layering — none of which live on a single flat struct the
// way the reference's config.Resolved does.
package cmdtest

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/analytics/analyticstest"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// RunIsolated sandboxes a package's tests with a temp XDG_CONFIG_HOME (so the
// token-refresh path never touches a developer's real auth.yaml) and a no-op
// analytics tracker (LogCommandParameters needs one set globally — see
// cli/app/main_test.go). Use from a package TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(cmdtest.RunIsolated(m)) }
//
// Every env var below reaches a command through a cli/cmdutil resolver rather
// than through config.WithResolved, so a developer who exports one would get
// different results than CI: an ambient BITRISE_WORKSPACE_ID alone silently
// redirects every request a test asserts a path for.
func RunIsolated(m *testing.M) int {
	dir, err := os.MkdirTemp("", "bitrise-rde-test")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := os.Setenv("XDG_CONFIG_HOME", dir); err != nil {
		panic(err)
	}
	for _, key := range []string{
		"BITRISE_TOKEN",
		cmdutil.EnvWorkspaceID,
		cmdutil.EnvAppID,
		cmdutil.EnvAppIDLegacy,
		cmdutil.EnvOutput,
		cmdutil.EnvTheme,
		cmdutil.EnvRDEAPIBaseURL,
	} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}
	cmdutil.SetTracker(analyticstest.NoOpTracker{})
	return m.Run()
}

// Opts configures a single Run of cmd under test.
type Opts struct {
	Args               []string
	RDEAPIBaseURL      string
	DefaultWorkspaceID string
	Format             string // "raw"/"json"/"yml"; empty leaves the command's own default
	Stdin              string
}

// Run executes cmd with opts and returns its captured stdout/stderr.
func Run(t *testing.T, cmd *cobra.Command, opts Opts) (stdout, stderr string, err error) {
	t.Helper()
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetIn(strings.NewReader(opts.Stdin))

	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{
		RDEAPIBaseURL:      opts.RDEAPIBaseURL,
		DefaultWorkspaceID: opts.DefaultWorkspaceID,
	})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))

	if opts.Format != "" {
		if err := cmd.Flags().Set(output.FormatKey, opts.Format); err != nil {
			t.Fatalf("set format flag: %v", err)
		}
	}

	cmd.SetArgs(opts.Args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}
