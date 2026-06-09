package indexgen

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/assert"
)

func TestIndex_per_step_latest_pointer(t *testing.T) {
	out := runGenerateFromSteplibClone(t)

	var latest steplibindex.LatestPointer
	readJSON(t, filepath.Join(out, "index/steps/hello-step/latest.json"), &latest)

	assert.Equal(t, "hello-step", latest.StepID, "StepID")
	assert.Equal(t, "2.0.0", latest.Latest, "Latest")
	assert.Equal(t, map[string]string{
		"1": "1.1.0",
		"2": "2.0.0",
	}, latest.LatestByMajor, "LatestByMajor")
}

func TestIndex_single_version_latest_pointer(t *testing.T) {
	out := runGenerateFromSteplibClone(t)

	// bash-step has exactly one version (1.0.0) in a single major.
	var latest steplibindex.LatestPointer
	readJSON(t, filepath.Join(out, "index/steps/bash-step/latest.json"), &latest)

	assert.Equal(t, "bash-step", latest.StepID, "StepID")
	assert.Equal(t, "1.0.0", latest.Latest, "Latest")
	assert.Equal(t, map[string]string{"1": "1.0.0"}, latest.LatestByMajor, "LatestByMajor")
}

func TestIndex_versions_newest_first(t *testing.T) {
	out := runGenerateFromSteplibClone(t)

	var versions steplibindex.Versions
	readJSON(t, filepath.Join(out, "index/steps/hello-step/versions.json"), &versions)

	assert.Equal(t, "hello-step", versions.StepID, "StepID")
	assert.Equal(t, []string{"2.0.0", "1.1.0", "1.0.0"}, versions.Versions, "Versions newest-first")
}
