//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Envman(t *testing.T) {
	ensureEnvmanInstalled(t)

	t.Run("lifecycle: init, add, print, json, clear", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		out, err := runBitriseEnvman(dir, "init", "--clear")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "add", "--key", "FOO", "--value", "bar")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "print")
		require.NoError(t, err, out)
		assert.Contains(t, out, "FOO")
		assert.Contains(t, out, "bar")

		// single key ⇒ stable JSON, so a substring match is safe.
		out, err = runBitriseEnvman(dir, "print", "--format", "json")
		require.NoError(t, err, out)
		assert.Contains(t, out, `"FOO":"bar"`)

		out, err = runBitriseEnvman(dir, "clear")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "print")
		require.NoError(t, err, out)
		assert.NotContains(t, out, "FOO")
	})

	t.Run("run injects stored env into the child process", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		out, err := runBitriseEnvman(dir, "init", "--clear")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "add", "--key", "GREETING", "--value", "hello")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "run", "bash", "-c", "echo $GREETING")
		require.NoError(t, err, out)
		assert.Contains(t, out, "hello")
	})

	t.Run("add reads the value from piped stdin", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		out, err := runBitriseEnvman(dir, "init", "--clear")
		require.NoError(t, err, out)

		out, err = command.New(testhelpers.BinPath(), "envman", "add", "--key", "PIPED").
			SetDir(dir).
			SetStdin(strings.NewReader("from-stdin")).
			RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "print")
		require.NoError(t, err, out)
		assert.Contains(t, out, "from-stdin")
	})

	t.Run("flags are passed through to envman (SkipFlagParsing)", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		out, err := runBitriseEnvman(dir, "--help")
		require.NoError(t, err, out)
		// These are not present in bitrise --help
		assert.Contains(t, out, "NAME: envman - Environment variable manager")
		assert.Contains(t, out, "--loglevel value, -l value")

		envstorePath := filepath.Join(dir, "custom.envstore.yml")
		out, err = runBitriseEnvman(dir, "--path", envstorePath, "init", "--clear")
		require.NoError(t, err, out)

		out, err = runBitriseEnvman(dir, "--path", envstorePath, "add", "--key", "PASSED", "--value", "through")
		require.NoError(t, err, out)
		assert.FileExists(t, envstorePath)

		out, err = runBitriseEnvman(dir, "--path", envstorePath, "print")
		require.NoError(t, err, out)
		assert.Contains(t, out, "PASSED")
	})

	t.Run("non-zero envman exit collapses to exit status 1", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		out, err := runBitriseEnvman(dir, "init", "--clear")
		require.NoError(t, err, out)

		// failf maps any envman failure to exit 1: the inner command exits 7, yet bitrise returns 1. "oops" confirms envman's stderr is piped through.
		out, err = runBitriseEnvman(dir, "run", "bash", "-c", "echo oops 1>&2; exit 7")
		require.EqualError(t, err, "exit status 1", out)
		assert.Contains(t, out, "oops")
	})
}

func ensureEnvmanInstalled(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "setup")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func runBitriseEnvman(dir string, args ...string) (string, error) {
	cmd := command.New(testhelpers.BinPath(), append([]string{"envman"}, args...)...)
	return cmd.SetDir(dir).RunAndReturnTrimmedCombinedOutput()
}
