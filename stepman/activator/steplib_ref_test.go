package activator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/stretchr/testify/require"
)

func Test_activateStepLibStep(t *testing.T) {
	tests := []struct {
		name        string
		stepIDData  stepid.CanonicalID
		wantVersion string
		wantErr     bool
	}{
		{
			name: "Major version lock",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1",
			},
			wantVersion: "1.10.1",
			wantErr:     false,
		},
		{
			name: "Major version lock (long form)",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1.x.x",
			},
			wantVersion: "1.10.1",
			wantErr:     false,
		},
		{
			name: "Minor version lock",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "2.3",
			},
			wantVersion: "2.3.7",
			wantErr:     false,
		},
		{
			name: "Minor version lock (long form)",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "2.3.x",
			},
			wantVersion: "2.3.7",
			wantErr:     false,
		},
		{
			name: "Patch version lock",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "2.3.2",
			},
			wantVersion: "2.3.2",
			wantErr:     false,
		},
		{
			name: "Invalid version lock",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1.2.3.4",
			},
			wantVersion: "",
			wantErr:     true,
		},
		{
			name: "Latest version (not supported at the moment)",
			stepIDData: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "x.x.x",
			},
			wantVersion: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := pathutil.NormalizedOSTempDirPath("activateStepLibStep")
			if err != nil {
				t.Errorf("failed to create tmp dir: %s", err)
			}
			stepYMLCopyPth := filepath.Join(tmpDir, "step-yml", "step.yml")

			if err := os.MkdirAll(filepath.Dir(stepYMLCopyPth), 0777); err != nil {
				t.Errorf("failed to create dir for step.yml: %s", err)
			}

			got, _, err := prepareStepLibForActivation(TestLogger[*testing.T]{t}, tt.stepIDData, false, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("activateStepLibStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Version != tt.wantVersion {
				t.Errorf("activateStepLibStep() got = %v, want %v", got.Version, tt.wantVersion)
			}
		})
	}
}

func TestActivateSteplibRefStep(t *testing.T) {
	const steplib = "https://github.com/bitrise-io/bitrise-steplib.git"
	logger := TestLogger[*testing.T]{t}

	tests := []struct {
		name        string
		id          stepid.CanonicalID
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "Exact version resolves and activates from source",
			id:          stepid.CanonicalID{SteplibSource: steplib, IDorURI: "xcode-archive", Version: "2.3.2"},
			wantVersion: "2.3.2",
		},
		{
			name:        "Minor version lock resolves to the highest patch",
			id:          stepid.CanonicalID{SteplibSource: steplib, IDorURI: "xcode-archive", Version: "2.3"},
			wantVersion: "2.3.7",
		},
		{
			name:    "Invalid version constraint fails",
			id:      stepid.CanonicalID{SteplibSource: steplib, IDorURI: "xcode-archive", Version: "1.2.3.4"},
			wantErr: true,
		},
		{
			name:    "Non-existent version fails",
			id:      stepid.CanonicalID{SteplibSource: steplib, IDorURI: "xcode-archive", Version: "99.99.99"},
			wantErr: true,
		},
		{
			name:    "Non-existent step fails",
			id:      stepid.CanonicalID{SteplibSource: steplib, IDorURI: "this-step-does-not-exist-xyz", Version: "1.0.0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BITRISE_STEPLIB_API_ENABLE", "false")

			activatedStepDir := t.TempDir()
			workDir := t.TempDir()

			// didStepLibUpdateInWorkflow=true keeps the StepLib update path off, so
			// resolution is served from the local cache and DidStepLibUpdate is false.
			got, err := ActivateSteplibRefStep(logger, tt.id, activatedStepDir, workDir, true, false)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.id.IDorURI, got.StepInfo.ID)
			require.Equal(t, tt.wantVersion, got.StepInfo.Version)
			require.Equal(t, ActivationTypeSteplibSource, got.ActivationType)
			require.Equal(t, ActivationInventorySourceSteplib, got.ActivationInventorySource)
			require.Empty(t, got.ExecutablePath)
			require.False(t, got.DidStepLibUpdate)

			require.Equal(t, filepath.Join(workDir, "current_step.yml"), got.StepYMLPath)
			exists, err := pathutil.IsPathExists(got.StepYMLPath)
			require.NoError(t, err)
			require.True(t, exists, "step.yml should be copied into the work dir")
		})
	}
}

// TestActivateSteplibRefStep_APIEnabled covers steplib API activation
// path. Each case runs in a fresh $HOME with the API flag on, so a passing run
// proves API resolves and activates against the hosted inventory without the
// git cloned steplib being set up
func TestActivateSteplibRefStep_APIEnabled(t *testing.T) {
	const steplib = "https://github.com/bitrise-io/bitrise-steplib.git"

	tests := []struct {
		name        string
		id          stepid.CanonicalID
		wantVersion string // exact concrete version; empty when only wantPrefix is checked
		wantPrefix  string
		wantErr     bool
	}{
		{
			name:        "Exact version activates from source over the API",
			id:          stepid.CanonicalID{SteplibSource: steplib, IDorURI: "git-clone", Version: "8.5.0"},
			wantVersion: "8.5.0",
		},
		{
			name:       "Minor version lock resolves to a concrete version",
			id:         stepid.CanonicalID{SteplibSource: steplib, IDorURI: "git-clone", Version: "8.4"},
			wantPrefix: "8.4.",
		},
		{
			name:    "Non-existent version fails",
			id:      stepid.CanonicalID{SteplibSource: steplib, IDorURI: "git-clone", Version: "99.99.99"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir()) // fresh: the git cloned steplib is not set up
			t.Setenv("BITRISE_STEPLIB_API_ENABLE", "true")
			t.Setenv("BITRISE_EXPERIMENT_PRECOMPILED_STEPS", "false")

			activatedStepDir := t.TempDir()
			workDir := t.TempDir()

			got, err := ActivateSteplibRefStep(TestLogger[*testing.T]{t}, tt.id, activatedStepDir, workDir, false, false)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.id.IDorURI, got.StepInfo.ID)
			require.Equal(t, tt.id.Version, got.StepInfo.OriginalVersion)
			if tt.wantVersion != "" {
				require.Equal(t, tt.wantVersion, got.StepInfo.Version)
			}
			if tt.wantPrefix != "" {
				require.True(t, strings.HasPrefix(got.StepInfo.Version, tt.wantPrefix),
					"resolved version %q should start with %q", got.StepInfo.Version, tt.wantPrefix)
			}
			require.Equal(t, ActivationTypeSteplibSource, got.ActivationType)
			require.Equal(t, ActivationInventorySourceSteplibAPI, got.ActivationInventorySource)
			require.Empty(t, got.ExecutablePath)
			require.False(t, got.DidStepLibUpdate)

			require.Equal(t, filepath.Join(workDir, "current_step.yml"), got.StepYMLPath)
			exists, err := pathutil.IsPathExists(got.StepYMLPath)
			require.NoError(t, err)
			require.True(t, exists, "current_step.yml should be written into the work dir")

			// API route must not have set up the git cloned steplib.
			_, found := stepman.ReadRoute(steplib)
			require.False(t, found, "v2 activation must not set up the v1 steplib")
		})
	}
}

type genericLogger interface {
	Logf(format string, v ...any)
}

type TestLogger[t genericLogger] struct {
	l genericLogger
}

func (t TestLogger[l]) Debugf(format string, v ...any) {
	t.l.Logf(format, v...)
}
func (t TestLogger[l]) Errorf(format string, v ...any) {
	t.l.Logf(format, v...)
}
func (t TestLogger[l]) Warnf(format string, v ...any) {
	t.l.Logf(format, v...)
}
func (t TestLogger[l]) Infof(format string, v ...any) {
	t.l.Logf(format, v...)
}

func BenchmarkActivateSteplibRefStep(b *testing.B) {
	logger := TestLogger[*testing.B]{b}
	tests := []struct {
		name                       string
		id                         stepid.CanonicalID
		isOfflineMode              bool
		didStepLibUpdateInWorkflow bool
		shouldCleanSteplib         bool
		wantErr                    bool
	}{
		{
			name: "No steplib update, major versiom",
			id: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1",
			},
			didStepLibUpdateInWorkflow: true,
			wantErr:                    false,
		},
		{
			name: "Steplib update, major versiom",
			id: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1",
			},
			didStepLibUpdateInWorkflow: false,
			wantErr:                    false,
		},
		{
			name: "Clean steplib every time",
			id: stepid.CanonicalID{
				SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git",
				IDorURI:       "xcode-archive",
				Version:       "1",
			},
			didStepLibUpdateInWorkflow: false,
			shouldCleanSteplib:         true,
			wantErr:                    false,
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				if tt.shouldCleanSteplib {
					err := os.RemoveAll("~/.stepman")
					require.NoError(b, err)
				}
				tmpDir, err := pathutil.NormalizedOSTempDirPath("activateStepLibStep")
				if err != nil {
					b.Errorf("failed to create tmp dir: %s", err)
				}
				stepYMLCopyPth := filepath.Join(tmpDir, "step-yml", "step.yml")

				if err := os.MkdirAll(filepath.Dir(stepYMLCopyPth), 0777); err != nil {
					b.Errorf("failed to create dir for step.yml: %s", err)
				}

				got, gotErr := ActivateSteplibRefStep(logger, tt.id, stepYMLCopyPth, tmpDir, tt.didStepLibUpdateInWorkflow, tt.isOfflineMode)
				if gotErr != nil {
					if !tt.wantErr {
						b.Errorf("ActivateSteplibRefStep() failed: %v", gotErr)
					}
					return
				}
				if tt.wantErr {
					b.Fatal("ActivateSteplibRefStep() succeeded unexpectedly")
				}

				require.Equal(b, tmpDir+"/current_step.yml", got.StepYMLPath)
				require.Equal(b, ActivationTypeSteplibSource, got.ActivationType)
				require.Equal(b, !tt.didStepLibUpdateInWorkflow, got.DidStepLibUpdate)
				require.Equal(b, tt.id.IDorURI, got.StepInfo.ID)
			}
		})
	}
}
