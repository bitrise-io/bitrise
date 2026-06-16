package indexgen

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerator_deterministic_output asserts that two generations with
// identical inputs and Options produce byte-identical output trees.
//
// This makes the regen-a-sample-tree workflow's implicit assumption an
// asserted contract. It also catches any future non-determinism (map
// iteration order leaking into output, time.Now() injected somewhere other
// than Options.GeneratedAt, etc.) before it confuses anyone diffing a
// regenerated sample tree.
//
// Comparison is by per-file SHA-256 keyed on the output-relative path
// (covering the whole v2/ tree); a failure prints which path's hash differs.
func TestGenerator_deterministic_output(t *testing.T) {
	// Two runs of the same helper = identical fixture, Options, and logger; any
	// output difference is therefore non-determinism in the generator.
	hashes1 := hashAllFiles(t, runGenerateFromSteplibClone(t))
	hashes2 := hashAllFiles(t, runGenerateFromSteplibClone(t))

	require.NotEmpty(t, hashes1, "first generation produced no files")
	assert.Equal(t, hashes1, hashes2,
		"two generations with identical Options must produce byte-identical trees")
}

// hashAllFiles walks dir and returns a map from output-relative path to the
// SHA-256 hex digest of that file's contents. Directories are skipped.
func hashAllFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	dirFS := os.DirFS(dir)
	out := map[string]string{}
	require.NoError(t, fs.WalkDir(dirFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		bytes, err := fs.ReadFile(dirFS, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(bytes)
		out[p] = hex.EncodeToString(sum[:])
		return nil
	}))
	return out
}
