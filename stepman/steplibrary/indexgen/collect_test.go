package indexgen

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalStepYAML returns the smallest step.yml body that passes Validate:
// a title plus a non-empty source block (git + a 40-char commit). Used by
// tests that build synthetic input filesystems and don't care about the
// step's content beyond "this version exists and is well-formed". Without a
// source block the generator's validate-before-publish step would reject the
// staged tree (missing source.git/source.commit).
func minimalStepYAML(title string) []byte {
	return []byte(
		"title: " + title + "\n" +
			"source:\n" +
			"  git: https://example.com/repo.git\n" +
			"  commit: '0000000000000000000000000000000000000000'\n",
	)
}

func TestCollect_step_info_and_asset_copy(t *testing.T) {
	root := runGenerateFromSteplibClone(t)

	var info steplibindex.StepInfo
	readJSON(t, filepath.Join(root, mustFS(steplibindex.StepInfoPath("hello-step"))), &info)

	assert.Equal(t, "bitrise", info.Maintainer, "Maintainer")
	assert.Nil(t, info.Deprecation, "Deprecation")
	assert.Equal(t, []string{"assets/icon.svg"}, info.AssetURLs, "AssetURLs")

	// Asset file copied.
	assert.FileExists(t, filepath.Join(root, mustFS(steplibindex.StepAssetPath("hello-step", "icon.svg"))), "asset file copied")
}

func TestCollect_asset_permissions_preserved(t *testing.T) {
	// Embedded fixtures can't carry real file permissions, so drive the
	// generator from an in-memory FS where the asset's mode is set explicitly to
	// a value distinct from common defaults (and from the 0600 the writer uses
	// for files it authors), then assert the copy preserves it.
	const mode = os.FileMode(0o640)
	inputFS := fstest.MapFS{
		"steplib.yml":                     {Data: []byte("format_version: '0.9.0'\nsteplib_source: 'https://example.com'\n")},
		"steps/perm-step/step-info.yml":   {Data: []byte("maintainer: test\n")},
		"steps/perm-step/1.0.0/step.yml":  {Data: minimalStepYAML("Perm Step")},
		"steps/perm-step/assets/icon.svg": {Data: []byte("<svg/>"), Mode: mode},
	}

	out := t.TempDir()
	_, gotErr := generateFromSteplibClone(inputFS, out, Options{GeneratedAt: fixedTime}, testLogger{t})
	require.NoError(t, gotErr, "generateFromSteplibClone")

	dstInfo, err := os.Stat(filepath.Join(out, mustFS(steplibindex.StepAssetPath("perm-step", "icon.svg"))))
	require.NoError(t, err, "stat copied asset")
	assert.Equal(t, mode, dstInfo.Mode().Perm(), "copied asset preserves source file mode")
}

func TestCollect_deprecated_step(t *testing.T) {
	root := runGenerateFromSteplibClone(t)

	var info steplibindex.StepInfo
	readJSON(t, filepath.Join(root, mustFS(steplibindex.StepInfoPath("deprecated-step"))), &info)

	assert.Equal(t, "bitrise", info.Maintainer, "Maintainer")
	require.NotNil(t, info.Deprecation, "Deprecation")
	assert.Equal(t, "2025-04-11", info.Deprecation.RemovalDate, "RemovalDate")
	assert.Contains(t, info.Deprecation.Notes, "key-based caching", "Notes")
	assert.Empty(t, info.AssetURLs, "no assets dir → no asset_urls")
}

func TestCollect_missing_step_info_is_error(t *testing.T) {
	// step-info.yml is mandatory: a step without one fails generation.
	inputFS := fstest.MapFS{
		"steplib.yml":                       {Data: []byte("format_version: '0.9.0'\n")},
		"steps/no-info-step/1.0.0/step.yml": {Data: []byte("title: No Info\n")},
	}
	out := t.TempDir()
	_, gotErr := generateFromSteplibClone(inputFS, out, Options{GeneratedAt: fixedTime}, testLogger{t})
	require.Error(t, gotErr, "missing step-info.yml must error")
	assert.ErrorIs(t, gotErr, fs.ErrNotExist, "error wraps fs.ErrNotExist")
	assert.Contains(t, gotErr.Error(), "no-info-step", "error names the offending step")
}

func TestCollect_invalid_version_dir_skipped(t *testing.T) {
	inputFS := fstest.MapFS{
		"steplib.yml":                  {Data: []byte("format_version: '0.9.0'\n")},
		"steps/my-step/step-info.yml":  {Data: []byte("maintainer: test\n")},
		"steps/my-step/1.0.0/step.yml": {Data: minimalStepYAML("My Step")},
		// Non-semver version dirs are skipped, so their content is never read.
		"steps/my-step/not-a-semver/step.yml":    {Data: []byte("title: Should be skipped\n")},
		"steps/my-step/also-not-semver/step.yml": {Data: []byte("title: Also skipped\n")},
	}
	out := t.TempDir()
	stats, gotErr := generateFromSteplibClone(inputFS, out, Options{GeneratedAt: fixedTime}, testLogger{t})
	require.NoError(t, gotErr, "generateFromSteplibClone")

	assert.Equal(t, 1, stats.StepCount, "StepCount")
	assert.Equal(t, 1, stats.VersionCount, "VersionCount")

	assert.FileExists(t, filepath.Join(out, mustFS(steplibindex.StepJSONPath("my-step", "1.0.0"))), "valid version written")
	assert.NoFileExists(t, filepath.Join(out, mustFS(steplibindex.StepJSONPath("my-step", "not-a-semver"))), "non-semver version dir skipped")
}

func TestCollect_multi_platform_executables(t *testing.T) {
	root := runGenerateFromSteplibClone(t)

	var step models.StepModel
	readJSON(t, filepath.Join(root, mustFS(steplibindex.StepJSONPath("multi-platform-step", "3.2.1"))), &step)

	require.NotNil(t, step.Executables, "Executables")
	require.Len(t, *step.Executables, 4, "Executables len")

	darwinArm := (*step.Executables)["darwin-arm64"]
	assert.Equal(t,
		"steps/deploy-to-bitrise-io/2.23.2/bin/deploy-to-bitrise-io-darwin-arm64",
		darwinArm.StorageURI,
		"darwin-arm64 StorageURI",
	)
	assert.Equal(t,
		"sha256-316b1ae22a53e06199b68a3ddf008345aa9e3690abcd57243085a56ccdc57159",
		darwinArm.Hash,
		"darwin-arm64 Hash",
	)

	linuxAmd := (*step.Executables)["linux-amd64"]
	assert.Equal(t,
		"steps/deploy-to-bitrise-io/2.23.2/bin/deploy-to-bitrise-io-linux-amd64",
		linuxAmd.StorageURI,
		"linux-amd64 StorageURI",
	)
}

func TestCollect_bash_step_has_no_executables(t *testing.T) {
	root := runGenerateFromSteplibClone(t)

	var step models.StepModel
	readJSON(t, filepath.Join(root, mustFS(steplibindex.StepJSONPath("bash-step", "1.0.0"))), &step)

	// The Script step ships no precompiled binary, so activation builds from
	// source (Executables nil). It also declares no toolkit, and the generator
	// preserves that verbatim — it never synthesizes a default. The bash +
	// step.sh default is applied at run time, not baked into step.json.
	assert.Nil(t, step.Executables, "Executables")
	assert.Nil(t, step.Toolkit, "Toolkit")
	require.NotNil(t, step.Title, "Title")
	assert.Equal(t, "Script", *step.Title, "Title")
}
