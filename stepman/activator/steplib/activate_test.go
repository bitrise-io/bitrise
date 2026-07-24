package steplib

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/steplibrary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for ActivateStep's dispatch: version resolution, step.yml placement, and
// the executable-vs-source-vs-fallback decision. Leaf mechanics — download URL
// building, mirror fallback, hash checks (activate_executable_test.go) and the
// source guard clauses (source_test.go) — are owned elsewhere; these tests assert
// only which branch ran and the shape of the result.

func TestActivateStep_ResolvesExactVersion(t *testing.T) {
	srv := serveHelloStepInventory(t)
	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "false")

	destination := t.TempDir()
	workDir := t.TempDir()
	stepYML := filepath.Join(workDir, "current_step.yml")

	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "2.0.0"}

	resolved, err := ActivateStep(id, destination, stepYML, log, false, steplibrary.New(log, srv.URL), &fakeExecutableFetcher{})
	require.NoError(t, err)

	assert.Equal(t, "hello-step", resolved.StepInfo.ID)
	assert.Equal(t, "2.0.0", resolved.StepInfo.Version, "exact request resolves to itself")
	assert.Equal(t, "2.0.0", resolved.StepInfo.OriginalVersion)
	assert.Empty(t, resolved.ExecPath, "source activation must not yield an executable path")

	require.FileExists(t, stepYML, "step.yml should be written to the work dir")
	require.FileExists(t, filepath.Join(destination, "activated_marker.txt"),
		"step source should be materialized into the destination dir")
}

func TestActivateStep_ResolvesMajorLock(t *testing.T) {
	srv := serveHelloStepInventory(t)
	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "false")

	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "1"}

	resolved, err := ActivateStep(id, t.TempDir(), filepath.Join(t.TempDir(), "current_step.yml"), log, false, steplibrary.New(log, srv.URL), &fakeExecutableFetcher{})
	require.NoError(t, err)

	assert.Equal(t, "1.1.0", resolved.StepInfo.Version, "major lock 1 resolves to the highest 1.x")
	assert.Equal(t, "1", resolved.StepInfo.OriginalVersion)
	assert.Empty(t, resolved.ExecPath)
}

func TestActivateStep_NonexistentVersionFails(t *testing.T) {
	srv := serveHelloStepInventory(t)
	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "false")

	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "99.99.99"}

	_, err := ActivateStep(id, t.TempDir(), filepath.Join(t.TempDir(), "current_step.yml"), log, false, steplibrary.New(log, srv.URL), &fakeExecutableFetcher{})
	require.Error(t, err, "a version not in the inventory must fail resolution")
}

// With the precompiled step flag on but no executable published for a step,
// ActivateStep falls back to source activation.
func TestActivateStep_NoExecutable_ActivatesSource(t *testing.T) {
	srv := serveHelloStepInventory(t)
	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "true")

	destination := t.TempDir()
	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "2.0.0"}

	resolved, err := ActivateStep(id, destination, filepath.Join(t.TempDir(), "current_step.yml"), log, false, steplibrary.New(log, srv.URL), &fakeExecutableFetcher{})
	require.NoError(t, err)

	assert.Empty(t, resolved.ExecPath, "a step without executables must fall back to source")
	require.FileExists(t, filepath.Join(destination, "activated_marker.txt"))
}

// With the precompiled step flag on and an executable published for the current
// platform, ActivateStep takes the executable branch and does not fall back to
// source. The binary transfer is faked; asserting the download URL/hash is the
// job of activate_executable_test.go.
func TestActivateStep_ChoosesExecutable(t *testing.T) {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	srv := serveHelloStepInventoryWithExecutables(t, &models.Executables{
		platform: models.Executable{
			StorageURI: "steps/hello-step/2.0.0/bin/hello-step-" + platform,
			Hash:       "sha256-1111111111111111111111111111111111111111111111111111111111111111",
		},
	})

	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "true")

	destination := t.TempDir()
	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "2.0.0"}

	resolved, err := ActivateStep(id, destination, filepath.Join(t.TempDir(), "current_step.yml"), log, false, steplibrary.New(log, srv.URL), &fakeExecutableFetcher{})
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(destination, "hello-step"), resolved.ExecPath)
	require.FileExists(t, resolved.ExecPath)
	require.NoFileExists(t, filepath.Join(destination, "activated_marker.txt"),
		"executable activation must not fall back to source")
}

// If the executable download fails, ActivateStep falls back to source activation
// rather than erroring.
func TestActivateStep_ExecutableDownloadFails_FallsBackToSource(t *testing.T) {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	srv := serveHelloStepInventoryWithExecutables(t, &models.Executables{
		platform: models.Executable{
			StorageURI: "steps/hello-step/2.0.0/bin/hello-step-" + platform,
			Hash:       "sha256-1111111111111111111111111111111111111111111111111111111111111111",
		},
	})

	log := apiTestLogger{t}
	t.Setenv(precompiledStepsEnv, "true")

	destination := t.TempDir()
	id := stepid.CanonicalID{SteplibSource: testSteplibURL, IDorURI: "hello-step", Version: "2.0.0"}

	fetcher := &fakeExecutableFetcher{downloadErr: errFakeDownload}
	resolved, err := ActivateStep(id, destination, filepath.Join(t.TempDir(), "current_step.yml"), log, false, steplibrary.New(log, srv.URL), fetcher)
	require.NoError(t, err, "a failed executable download must fall back to source, not error")

	assert.Empty(t, resolved.ExecPath)
	require.FileExists(t, filepath.Join(destination, "activated_marker.txt"))
}
