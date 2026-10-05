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
			adjust: func(e *Eligibility) { e.Command, e.LogFormat = "run", "json" },
		},
		{
			name:   "console log format",
			adjust: func(e *Eligibility) { e.Command, e.LogFormat = "run", "console" },
			want:   true,
		},
		{
			name:   "the update command",
			adjust: func(e *Eligibility) { e.Command = updateCommandName },
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
			e := Eligibility{
				CurrentVersion: "2.45.0",
				Command:        "list",
				StderrIsTTY:    true,
				LookupEnv:      noEnv,
			}
			if tt.adjust != nil {
				tt.adjust(&e)
			}

			require.Equal(t, tt.want, IsEligible(e))
		})
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
