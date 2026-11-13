package stepid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StepIDData_IsUniqueResourceID(t *testing.T) {
	stepIDDataWithIDAndVersionSpecified := CanonicalID{IDorURI: "stepid", Version: "version"}
	stepIDDataWithOnlyVersionSpecified := CanonicalID{Version: "version"}
	stepIDDataWithOnlyIDSpecified := CanonicalID{IDorURI: "stepid"}
	stepIDDataEmpty := CanonicalID{}

	// Not Unique
	for _, aSourceID := range []string{"path", "git", "_", ""} {
		stepIDDataWithIDAndVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithIDAndVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyIDSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyIDSpecified.IsUniqueResourceID())

		stepIDDataEmpty.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataEmpty.IsUniqueResourceID())
	}

	for _, aSourceID := range []string{"a", "any-other-step-source", "https://github.com/bitrise-io/bitrise-steplib.git"} {
		// Only if StepLib, AND both ID and Version are defined, only then
		// this is a Unique Resource ID!
		stepIDDataWithIDAndVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, true, stepIDDataWithIDAndVersionSpecified.IsUniqueResourceID())

		// In any other case, it's not,
		// even if it's from a StepLib
		// but missing ID or version!
		stepIDDataWithOnlyVersionSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyVersionSpecified.IsUniqueResourceID())

		stepIDDataWithOnlyIDSpecified.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataWithOnlyIDSpecified.IsUniqueResourceID())

		stepIDDataEmpty.SteplibSource = aSourceID
		require.Equal(t, false, stepIDDataEmpty.IsUniqueResourceID())
	}
}

func TestCreateCanonicalIDFromString(t *testing.T) {
	type data struct {
		composite            string
		defaultSteplibSource string
		wantStepSrc          string
		wantStepID           string
		wantVersion          string
		wantErr              bool
		name                 string
	}

	stepLib := []data{
		{
			name:      "no steplib-source",
			composite: "step-id@0.0.1", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "invalid/empty step lib source, but default provided",
			composite: "::step-id@0.0.1", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "invalid/empty step lib source, no default",
			composite: "::step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "no steplib-source & no default, fail",
			composite: "step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "no steplib, no version, only step-id",
			composite: "step-id", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "default-steplib-src", wantStepID: "step-id", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "default, long, verbose ID mode",
			composite: "steplib-src::step-id@0.0.1", defaultSteplibSource: "",
			wantStepSrc: "steplib-src", wantStepID: "step-id", wantVersion: "0.0.1",
			wantErr: false,
		},
		{
			name:      "empty test",
			composite: "", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "special empty test",
			composite: "@1.0.0", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "", wantStepID: "", wantVersion: "",
			wantErr: true,
		},
		{
			name:      "old step",
			composite: "_::https://github.com/bitrise-io/steps-timestamp.git@1.0.0", defaultSteplibSource: "",
			wantStepSrc: "_", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "1.0.0",
			wantErr: false,
		},
	}

	path := []data{
		{
			name:      "local path",
			composite: "path::/some/path", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "/some/path", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "local path, tilde",
			composite: "path::~/some/path/in/home", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "~/some/path/in/home", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "local path, env",
			composite: "path::$HOME/some/path/in/home", defaultSteplibSource: "",
			wantStepSrc: "path", wantStepID: "$HOME/some/path/in/home", wantVersion: "",
			wantErr: false,
		},
	}

	git := []data{
		{
			name:      "direct git uri, https",
			composite: "git::https://github.com/bitrise-io/steps-timestamp.git@develop", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "develop",
			wantErr: false,
		},
		{
			name:      "direct git uri, ssh",
			composite: "git::git@github.com:bitrise-io/steps-timestamp.git@develop", defaultSteplibSource: "",
			wantStepSrc: "git", wantStepID: "git@github.com:bitrise-io/steps-timestamp.git", wantVersion: "develop",
			wantErr: false,
		},
		{
			name:      "direct git uri, https, no branch",
			composite: "git::https://github.com/bitrise-io/steps-timestamp.git", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "https://github.com/bitrise-io/steps-timestamp.git", wantVersion: "",
			wantErr: false,
		},
		{
			name:      "direct git uri, ssh, no branch",
			composite: "git::git@github.com:bitrise-io/steps-timestamp.git", defaultSteplibSource: "default-steplib-src",
			wantStepSrc: "git", wantStepID: "git@github.com:bitrise-io/steps-timestamp.git", wantVersion: "",
			wantErr: false,
		},
	}

	for _, group := range [][]data{
		stepLib,
		git,
		path,
	} {
		for _, tt := range group {
			stepIDData, err := CreateCanonicalIDFromString(tt.composite, tt.defaultSteplibSource)

			if tt.wantErr && (err == nil) {
				t.Fatal("tt.wantErr && (err == nil):", err)
			}

			require.Equal(t, tt.wantStepSrc, stepIDData.SteplibSource, tt.name)
			require.Equal(t, tt.wantStepID, stepIDData.IDorURI, tt.name)
			require.Equal(t, tt.wantVersion, stepIDData.Version, tt.name)
		}
	}
}
