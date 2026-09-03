package workarounds

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestShouldEnableRubyPrecompiled(t *testing.T) {
	tests := []struct {
		name        string
		toolName    provider.ToolID
		miseVersion string
		want        bool
	}{
		{
			name:        "Ruby with the stable pin (precompiled gated behind experimental) - should set",
			toolName:    "ruby",
			miseVersion: "v2026.5.12",
			want:        true,
		},
		{
			name:        "Ruby with a mise version where precompiled is the default - should not set",
			toolName:    "ruby",
			miseVersion: "v2026.8.0",
			want:        false,
		},
		{
			name:        "Ruby with a newer mise version - should not set",
			toolName:    "ruby",
			miseVersion: "v2026.9.1",
			want:        false,
		},
		{
			name:        "Ruby with a mise version predating the feature - should set, harmless no-op",
			toolName:    "ruby",
			miseVersion: "v2026.3.16",
			want:        true,
		},
		{
			name:        "nixpkgs-prefixed Ruby - should set",
			toolName:    "nixpkgs:ruby",
			miseVersion: "v2026.5.12",
			want:        true,
		},
		{
			name:        "Another tool - should not set",
			toolName:    "nodejs",
			miseVersion: "v2026.5.12",
			want:        false,
		},
		{
			name:        "Python is unaffected - should not set",
			toolName:    "python",
			miseVersion: "v2026.5.12",
			want:        false,
		},
		{
			name:        "Mise version without the v prefix - should set",
			toolName:    "ruby",
			miseVersion: "2026.5.12",
			want:        true,
		},
		{
			name:        "Unparseable mise version - should not set",
			toolName:    "ruby",
			miseVersion: "not-a-version",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldEnableRubyPrecompiled(tt.toolName, tt.miseVersion)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetRubyPrecompiledEnv(t *testing.T) {
	key, value := GetRubyPrecompiledEnv("ruby", "v2026.5.12", true)
	require.Equal(t, "MISE_EXPERIMENTAL", key)
	require.Equal(t, "true", value)

	key, value = GetRubyPrecompiledEnv("ruby", "v2026.9.1", true)
	require.Empty(t, key)
	require.Empty(t, value)

	key, value = GetRubyPrecompiledEnv("nodejs", "v2026.5.12", true)
	require.Empty(t, key)
	require.Empty(t, value)
}
