package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	internalconfig "github.com/bitrise-io/bitrise/v3/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSaveState_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	state := State{
		CheckedAt:      time.Now().UTC().Truncate(time.Second),
		CheckedVersion: "2.45.0",
		Versions: Versions{
			Update:   available("2.46.0"),
			NewMajor: available("3.0.0"),
		},
	}

	require.NoError(t, SaveState(state))

	require.Equal(t, state, LoadState())
}

func TestSaveState_LeavesTheConfigFileAlone(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	require.NoError(t, SaveState(State{CheckedVersion: "2.45.0"}))

	configPath, err := internalconfig.Path()
	require.NoError(t, err)
	require.NoFileExists(t, configPath)
}

func TestLoadState_UnusableFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		write   bool
	}{
		{name: "no state file yet"},
		{name: "unparsable state file", content: "checked_at: [", write: true},
		{name: "empty state file", content: "", write: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)
			if tt.write {
				path := filepath.Join(dir, "bitrise", "cli", stateFileName)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
				require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o600))
			}

			require.Equal(t, State{}, LoadState())
		})
	}
}

func TestState_NeedsCheck(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{
			name:  "checked for this version within the interval",
			state: State{CheckedAt: now.Add(-time.Hour), CheckedVersion: "2.45.0"},
		},
		{
			name:  "checked for this version before the interval",
			state: State{CheckedAt: now.Add(-checkInterval), CheckedVersion: "2.45.0"},
			want:  true,
		},
		{
			name:  "checked for another version",
			state: State{CheckedAt: now, CheckedVersion: "2.44.0"},
			want:  true,
		},
		{
			name: "never checked",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.state.NeedsCheck(now, "2.45.0"))
		})
	}
}
