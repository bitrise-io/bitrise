package update

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEligible(t *testing.T) {
	tests := []struct {
		name   string
		adjust func(e *Eligibility)
		want   bool
	}{
		{
			name: "an interactive run of a comparable build",
			want: true,
		},
		{
			name:   "a build with no release version",
			adjust: func(e *Eligibility) { e.CurrentVersion = "dev" },
		},
		{
			name:   "a snapshot build",
			adjust: func(e *Eligibility) { e.CurrentVersion = "2.46.0-next" },
		},
		{
			name:   "CI mode",
			adjust: func(e *Eligibility) { e.CIMode = true },
		},
		{
			name:   "quiet",
			adjust: func(e *Eligibility) { e.Quiet = true },
		},
		{
			name:   "stderr is not a terminal",
			adjust: func(e *Eligibility) { e.StderrIsTTY = false },
		},
		{
			name:   "JSON output",
			adjust: func(e *Eligibility) { e.OutputFormat = "json" },
		},
		{
			name:   "YAML output",
			adjust: func(e *Eligibility) { e.OutputFormat = "yml" },
		},
		{
			name:   "raw output",
			adjust: func(e *Eligibility) { e.OutputFormat = "raw" },
			want:   true,
		},
		{
			name:   "JSON log format",
			adjust: func(e *Eligibility) { e.CommandPath, e.LogFormat = "bitrise local run", "json" },
		},
		{
			name:   "console log format",
			adjust: func(e *Eligibility) { e.CommandPath, e.LogFormat = "bitrise local run", "console" },
			want:   true,
		},
		{
			name:   "opted out",
			adjust: func(e *Eligibility) { e.LookupEnv = envWith(EnvNoUpdateNotifier, "1") },
		},
		{
			name:   "opted out with the CLI-prefixed variable",
			adjust: func(e *Eligibility) { e.LookupEnv = envWith(EnvNoCLIUpdateNotifier, "") },
		},
		{
			name:   "an unrelated variable",
			adjust: func(e *Eligibility) { e.LookupEnv = envWith("BITRISE_NO_UPDATE", "1") },
			want:   true,
		},
		{
			name:   "no env lookup",
			adjust: func(e *Eligibility) { e.LookupEnv = nil },
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eligibleRun()
			if tt.adjust != nil {
				tt.adjust(&e)
			}

			require.Equal(t, tt.want, IsEligible(e))
		})
	}
}

// TestIsEligible_UpdateCommands lists every command path of the CLI that ends in
// `update`: only the CLI's own update says nothing about a new version itself.
func TestIsEligible_UpdateCommands(t *testing.T) {
	tests := []struct {
		commandPath string
		want        bool
	}{
		{commandPath: "bitrise update"},
		{commandPath: "bitrise yml update", want: true},
		{commandPath: "bitrise plugin update", want: true},
		{commandPath: "bitrise rde template update", want: true},
		{commandPath: "bitrise rde saved-input update", want: true},
		{commandPath: "bitrise rde session update", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.commandPath, func(t *testing.T) {
			e := eligibleRun()
			e.CommandPath = tt.commandPath

			require.Equal(t, tt.want, IsEligible(e))
		})
	}
}

// TestIsEligible_RenamedBinary pins that the rule reads the command, not the
// name the binary was installed under.
func TestIsEligible_RenamedBinary(t *testing.T) {
	e := eligibleRun()
	e.CommandPath = "bitrise-cli update"

	require.False(t, IsEligible(e))
}

func eligibleRun() Eligibility {
	return Eligibility{
		CurrentVersion: "2.45.0",
		CommandPath:    "bitrise build list",
		StderrIsTTY:    true,
		LookupEnv:      noEnv,
	}
}

func noEnv(string) (string, bool) {
	return "", false
}

func envWith(key, value string) func(string) (string, bool) {
	return func(lookup string) (string, bool) {
		if lookup == key {
			return value, true
		}
		return "", false
	}
}
