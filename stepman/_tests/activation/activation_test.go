//go:build integration

package activation

import (
	"strings"
	"testing"

	"github.com/bitrise-io/stepman/activator"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertStepInfoContent checks the resolved StepInfo carries the requested
// identity, echoes the requested version as OriginalVersion, and holds a
// concrete Version consistent with the requested form. It runs against both the
// v1 and v2 paths, so it only asserts fields both populate (not Library or
// DefinitionPth, which differ by path). form is the version-form label.
func assertStepInfoContent(t *testing.T, variantName string, info models.StepInfoModel, id stepid.CanonicalID, form string) {
	t.Helper()
	assert.Equal(t, id.IDorURI, info.ID, "%s: StepInfo.ID", variantName)
	assert.Equal(t, id.Version, info.OriginalVersion, "%s: StepInfo.OriginalVersion echoes the requested version", variantName)
	assert.NotEmpty(t, info.Version, "%s: StepInfo.Version (concrete) must be set", variantName)
	assert.NotEmpty(t, info.LatestVersion, "%s: StepInfo.LatestVersion must be set", variantName)
	require.NotNil(t, info.Step.Title, "%s: StepInfo.Step must be populated (Title present)", variantName)
	assert.NotEmpty(t, *info.Step.Title, "%s: StepInfo.Step.Title", variantName)

	switch form {
	case "exact":
		assert.Equal(t, id.Version, info.Version, "%s: exact request resolves to itself", variantName)
	case "minor-lock", "major-lock":
		assert.True(t, strings.HasPrefix(info.Version, id.Version+"."),
			"%s: %s of %q should resolve within it, got %q", variantName, form, id.Version, info.Version)
	}
}

// TestSteplibActivation drives activator.ActivateSteplibRefStep across the v1/v2
// matrix against real steplib URLs. It runs in a hermetic $HOME so the first v1
// activation clones bitrise-steplib into a throwaway ~/.stepman and the real
// cache is never touched. Each case logs its verdict, raw logs and timing;
// paired cases print v1 then v2 adjacently for an eyeball diff.
//
// Not parallel: it relies on process-global env (HOME + experiment flags) set
// via t.Setenv.
func TestSteplibActivation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	allVariants := []variant{v1Source, v1Precompiled, v2Source, v2Precompiled}

	// 1. Existing-version activation across version forms and all four variants.
	t.Run("version-forms", func(t *testing.T) {
		versions := []struct{ label, version string }{
			{"latest", ""},
			{"exact", "8.5.0"},
			{"minor-lock", "8.4"},
			{"major-lock", "8"},
		}
		for _, ver := range versions {
			ver := ver
			t.Run(ver.label, func(t *testing.T) {
				id := steplibStep("git-clone", ver.version)
				t.Logf("=== git-clone @ %q (%s) ===", ver.version, ver.label)
				for _, v := range allVariants {
					r := activate(t, v, id, false, false)
					logResult(t, v.name, r)
					require.NoError(t, r.err, "%s should activate git-clone@%q", v.name, ver.version)
					assertStepInfoContent(t, v.name, r.activated.StepInfo, id, ver.label)
				}
			})
		}
	})

	// 2. Precompiled vs source vs fallback-to-source.
	t.Run("precompiled-source-fallback", func(t *testing.T) {
		t.Run("go-step-precompiled-gets-executable", func(t *testing.T) {
			id := steplibStep("git-clone", "8.5.0")
			for _, v := range []variant{v1Precompiled, v2Precompiled} {
				r := activate(t, v, id, false, false)
				logResult(t, v.name, r)
				require.NoError(t, r.err)
				assert.Equal(t, activator.ActivationTypeSteplibExecutable, r.activated.ActivationType,
					"%s of a Go step with a prebuilt binary should be an executable activation", v.name)
			}
		})
		t.Run("bash-step-precompiled-falls-back-to-source", func(t *testing.T) {
			id := steplibStep("script", "1.2.1")
			for _, v := range []variant{v1Precompiled, v2Precompiled} {
				r := activate(t, v, id, false, false)
				logResult(t, v.name, r)
				require.NoError(t, r.err)
				assert.Equal(t, activator.ActivationTypeSteplibSource, r.activated.ActivationType,
					"%s of a bash step (no binary) must fall back to source", v.name)
			}
		})
		t.Run("version-without-binary-falls-back-to-source", func(t *testing.T) {
			id := steplibStep("git-clone", "8.4.0") // 8.4.0 ships no prebuilt binary
			r := activate(t, v2Precompiled, id, false, false)
			logResult(t, v2Precompiled.name, r)
			require.NoError(t, r.err)
			assert.Equal(t, activator.ActivationTypeSteplibSource, r.activated.ActivationType)
		})
		t.Run("source-fetch", func(t *testing.T) {
			id := steplibStep("git-clone", "8.5.0")
			for _, v := range []variant{v1Source, v2Source} {
				r := activate(t, v, id, false, false)
				logResult(t, v.name, r)
				require.NoError(t, r.err)
				assert.Equal(t, activator.ActivationTypeSteplibSource, r.activated.ActivationType)
			}
		})
	})

	// 3. v1 source cache: cold (miss → download) then warm (hit).
	t.Run("cache-cold-warm", func(t *testing.T) {
		id := steplibStep("git-clone", "8.5.0")
		evictFromCache(t, id)
		cold := activate(t, v1Source, id, false, false)
		logResult(t, "v1-source COLD (cache miss)", cold)
		require.NoError(t, cold.err)
		warm := activate(t, v1Source, id, false, false)
		logResult(t, "v1-source WARM (cache hit)", warm)
		require.NoError(t, warm.err)
		t.Logf("cold=%s warm=%s (warm should skip the source download)", cold.elapsed.Round(1e6), warm.elapsed.Round(1e6))
	})

	// 4. Offline mode (v1): warmed version succeeds, non-cached version errors.
	t.Run("offline", func(t *testing.T) {
		warmed := steplibStep("git-clone", "8.5.0")
		require.NoError(t, activate(t, v1Source, warmed, false, false).err, "prime the cache online")
		hit := activate(t, v1Source, warmed, true, false)
		logResult(t, "offline + cached", hit)
		assert.NoError(t, hit.err, "offline activation of a cached version should succeed")

		missing := steplibStep("git-clone", "8.4.1")
		evictFromCache(t, missing)
		miss := activate(t, v1Source, missing, true, false)
		logResult(t, "offline + not cached", miss)
		assert.Error(t, miss.err, "offline activation of a non-cached version should fail")
	})

	// 5. Error cases — paired v1 vs v2, both must fail; capture messages.
	t.Run("errors", func(t *testing.T) {
		cases := []struct {
			name    string
			id, ver string
		}{
			{"missing-step-id", "", "1.0.0"},
			{"invalid-step-id", "no-such-step-xyz", "1.0.0"},
			{"invalid-version-constraint", "git-clone", "not-a-version"},
			{"literal-latest-not-a-constraint", "git-clone", "latest"},
			{"exact-version-not-found", "git-clone", "99.99.99"},
			{"minor-lock-not-found", "git-clone", "8.99"},
			{"major-lock-not-found", "git-clone", "99"},
		}
		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				id := steplibStep(c.id, c.ver)
				r1, r2 := logPair(t, id, v1Source, v2Source, false)
				assert.Error(t, r1.err, "v1 should fail: %s", c.name)
				assert.Error(t, r2.err, "v2 should fail: %s", c.name)
			})
		}
	})

	// 6. didStepLibUpdateInWorkflow flag (v1): true skips the steplib update.
	t.Run("didsteplibupdate-flag", func(t *testing.T) {
		id := steplibStep("git-clone", "8.5.0")
		off := activate(t, v1Source, id, false, false)
		logResult(t, "didStepLibUpdateInWorkflow=false", off)
		require.NoError(t, off.err)
		on := activate(t, v1Source, id, false, true)
		logResult(t, "didStepLibUpdateInWorkflow=true", on)
		require.NoError(t, on.err)
	})
}
