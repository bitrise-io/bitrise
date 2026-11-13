package indexgen

import (
	"cmp"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGenerate_local is a manual, throwaway end-to-end generation against a
// real steplib, for eyeballing the output tree locally. It is skipped unless
// RUN_STEPLIB_GEN is set, so it never runs in CI / a normal `go test ./...`:
//
//	RUN_STEPLIB_GEN=1 go test -run TestGenerate_local -v ./steplibrary/indexgen/
//
// STEPLIB_URI overrides the source (default: bitrise-steplib). STEPLIB_OUT
// overrides the output dir; by default a fresh temp dir is used and left in
// place (logged) so you can inspect <out>/v2/.
func TestGenerate_local(t *testing.T) {
	if os.Getenv("RUN_STEPLIB_GEN") == "" {
		t.Skip("set RUN_STEPLIB_GEN=1 to generate against a real steplib")
	}

	uri := cmp.Or(os.Getenv("STEPLIB_URI"), "https://github.com/bitrise-io/bitrise-steplib.git")
	out := os.Getenv("STEPLIB_OUT")
	if out == "" {
		var err error
		out, err = os.MkdirTemp("", "steplib-v2-*")
		require.NoError(t, err, "create output dir")
	}

	// Options left zero: GeneratedAt defaults to now, SteplibCommitSHA to HEAD.
	stats, err := Generate(uri, out, Options{}, testLogger{t})
	require.NoError(t, err, "Generate")
	t.Logf("generated %d steps / %d versions -> %s/v2", stats.StepCount, stats.VersionCount, out)
}
