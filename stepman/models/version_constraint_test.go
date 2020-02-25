package models

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/google/go-cmp/cmp"
)

func Test_parseRequiredVersion(t *testing.T) {
	tests := []struct {
		name            string
		requiredVersion string
		want            VersionConstraint
		wantErr         bool
	}{
		{
			name:            "fixed version",
			requiredVersion: "1.2.3",
			want: VersionConstraint{
				VersionLockType: Fixed,
				Version: Semver{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
		},
		{
			name:            "locked minor version",
			requiredVersion: "1.2.x",
			want: VersionConstraint{
				VersionLockType: MinorLocked,
				Version: Semver{
					Major: 1,
					Minor: 2,
				},
			},
		},
		{
			name:            "locked minor version (omitted)",
			requiredVersion: "1.2",
			want: VersionConstraint{
				VersionLockType: MinorLocked,
				Version: Semver{
					Major: 1,
					Minor: 2,
				},
			},
		},
		{
			name:            "locked minor version (omitted) - Invalid due to trailing dot",
			requiredVersion: "1.2.",
			want:            VersionConstraint{},
			wantErr:         true,
		},
		{
			name:            "locked major version",
			requiredVersion: "1.x.x",
			want: VersionConstraint{
				VersionLockType: MajorLocked,
				Version: Semver{
					Major: 1,
				},
			},
		},
		{
			name:            "locked major version - Invalid due to omitted patch version",
			requiredVersion: "1.x",
			want:            VersionConstraint{},
			wantErr:         true,
		},
		{
			name:            "locked major version (omitted)",
			requiredVersion: "1",
			want: VersionConstraint{
				VersionLockType: MajorLocked,
				Version: Semver{
					Major: 1,
				},
			},
		},
		{
			name:            "locked major version (omitted) - Invalid due to trailing dot",
			requiredVersion: "1.",
			want:            VersionConstraint{},
			wantErr:         true,
		},
		{
			name:            "locked major version (omitted) - Invalid due to omitted minor version",
			requiredVersion: "1..x",
			want:            VersionConstraint{},
			wantErr:         true,
		},
		{
			name:            "Invalid major version, x",
			requiredVersion: "x.x.x",
			want:            VersionConstraint{},
			wantErr:         true,
		},
		{
			name:            "No version specified",
			requiredVersion: "",
			want: VersionConstraint{
				VersionLockType: Latest,
			},
			wantErr: false,
		},
		{
			name:            "Invalid major version, negative number",
			requiredVersion: "-1",
			want:            VersionConstraint{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequiredVersion(tt.requiredVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequiredVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRequiredVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_latestMatchingStepVersion(t *testing.T) {
	step := StepModel{
		Title: pointers.NewStringPtr("name 1"),
	}
	stepGroup := StepGroupModel{
		Versions: map[string]StepModel{
			"1.0.0": step,
			"1.1.0": step,
			"1.1.1": step,
			"1.2.0": step,
			"2.0.0": step,
			"2.1.1": step,
		},
		LatestVersionNumber: "2.0.0",
	}

	tests := []struct {
		name            string
		requiredVersion VersionConstraint
		stepVersions    StepGroupModel
		want            StepVersionModel
		want1           bool
	}{
		{
			name: "Fix version",
			requiredVersion: VersionConstraint{
				VersionLockType: Fixed,
				Version: Semver{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
			},
			stepVersions: stepGroup,
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.0.0",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
		},
		{
			name: "Lock Minor version",
			requiredVersion: VersionConstraint{
				VersionLockType: MinorLocked,
				Version: Semver{
					Major: 1,
					Minor: 1,
				},
			},
			stepVersions: stepGroup,
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.1.1",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
		},
		{
			name: "Lock Major ersion",
			requiredVersion: VersionConstraint{
				VersionLockType: MajorLocked,
				Version: Semver{
					Major: 1,
				},
			},
			stepVersions: stepGroup,
			want: StepVersionModel{
				Step:                   step,
				Version:                "1.2.0",
				LatestAvailableVersion: "2.0.0",
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := latestMatchingStepVersion(tt.requiredVersion, tt.stepVersions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("latestMatchingRequiredVersion() got = %+v, want +%v,\n Diff: %s", got, tt.want, cmp.Diff(tt.want1, got))
			}
			if got1 != tt.want1 {
				t.Errorf("latestMatchingRequiredVersion() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
