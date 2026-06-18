package steplibrary

import (
	"testing"

	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/steplibrary/steplibindex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToStepGroupInfoModel(t *testing.T) {
	cases := map[string]struct {
		given steplibindex.StepInfo
		want  models.StepGroupInfoModel
	}{
		"active step has empty deprecation fields": {
			given: steplibindex.StepInfo{
				Maintainer:  "bitrise",
				Deprecation: nil,
				AssetURLs:   []string{"assets/icon.svg"},
			},
			want: models.StepGroupInfoModel{
				Maintainer:     "bitrise",
				AssetURLs:      map[string]string{"icon.svg": "assets/icon.svg"},
				RemovalDate:    "",
				DeprecateNotes: "",
			},
		},
		"deprecated step flattens nested fields": {
			given: steplibindex.StepInfo{
				Maintainer: "community",
				Deprecation: &steplibindex.Deprecation{
					RemovalDate: "2025-12-31",
					Notes:       "Replaced by `new-step`.",
				},
				AssetURLs: nil,
			},
			want: models.StepGroupInfoModel{
				Maintainer:     "community",
				AssetURLs:      nil,
				RemovalDate:    "2025-12-31",
				DeprecateNotes: "Replaced by `new-step`.",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := toStepGroupInfoModel(tc.given)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSteplib_getStepVersionInfo(t *testing.T) {
	// newFakeAPI's "script" step exposes versions 1.0.0, 1.1.5, 1.2.0, 2.0.0,
	// 2.4.0, 2.4.1, 3.0.0 with latest 3.0.0 and latest-by-major 1→1.2.0,
	// 2→2.4.1, 3→3.0.0.
	cases := map[string]struct {
		stepID      string
		version     string
		wantVersion string
		wantErr     bool
	}{
		"latest resolves to newest version":  {stepID: "script", version: "", wantVersion: "3.0.0"},
		"fixed version that exists":          {stepID: "script", version: "2.4.1", wantVersion: "2.4.1"},
		"fixed version that is missing":      {stepID: "script", version: "2.4.9", wantErr: true},
		"major-locked picks latest of major": {stepID: "script", version: "2", wantVersion: "2.4.1"},
		"minor-locked picks highest patch":   {stepID: "script", version: "1.1", wantVersion: "1.1.5"},
		"major-locked with no such major":    {stepID: "script", version: "9", wantErr: true},
		"unknown step id":                    {stepID: "nope", version: "1.0.0", wantErr: true},
		"empty step id":                      {stepID: "", version: "1.0.0", wantErr: true},
		"invalid version constraint":         {stepID: "script", version: "1.2.3.4", wantErr: true},
	}

	client := &Client{
		log:         nil,
		steplibURI:  "https://steplib.example",
		api:         newFakeAPI(),
		fileManager: nil,
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			stepInfo, resolved, gotErr := client.getStepVersionInfo(t.Context(), tc.stepID, tc.version)
			if tc.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tc.wantVersion, resolved.Version, "resolved version")
			assert.Equal(t, tc.stepID, resolved.ID, "resolved id")
			assert.Equal(t, tc.wantVersion, stepInfo.Version, "step info version")
			assert.Equal(t, tc.version, stepInfo.OriginalVersion, "step info original version")
			assert.Equal(t, "3.0.0", stepInfo.LatestVersion, "step info latest version")
			assert.Equal(t, "https://steplib.example", stepInfo.Library, "step info library")
		})
	}
}

func TestResolveMinorLocked(t *testing.T) {
	cases := map[string]struct {
		versions []string
		major    uint64
		minor    uint64
		want     string
		wantErr  bool
	}{
		"picks the highest patch in the minor": {versions: []string{"1.1.0", "1.1.5", "1.2.0"}, major: 1, minor: 1, want: "1.1.5"},
		"single matching version":              {versions: []string{"2.4.1"}, major: 2, minor: 4, want: "2.4.1"},
		"no version matches the minor":         {versions: []string{"1.0.0", "2.0.0"}, major: 1, minor: 5, wantErr: true},
		"unparseable version is an error":      {versions: []string{"not-semver"}, major: 1, minor: 0, wantErr: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, gotErr := resolveMinorLocked(tc.versions, models.Semver{Major: tc.major, Minor: tc.minor, Patch: 0})
			if tc.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tc.want, got, "resolved version")
		})
	}
}
