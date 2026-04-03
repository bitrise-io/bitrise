package workarounds

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestShouldSetPythonPrecompiledFlavor(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		concreteVersion string
		miseVersion     string
		want            bool
	}{
		{
			name:            "Python 3.14 with mise 2025 - should set",
			toolName:        "python",
			concreteVersion: "3.14.0",
			miseVersion:     "v2025.12.1",
			want:            true,
		},
		{
			name:            "Python 3.14 with mise 2026 - should not set",
			toolName:        "python",
			concreteVersion: "3.14.0",
			miseVersion:     "v2026.3.10",
			want:            false,
		},
		{
			name:            "Python 3.13 with mise 2025 - should not set",
			toolName:        "python",
			concreteVersion: "3.13.0",
			miseVersion:     "v2025.12.1",
			want:            false,
		},
		{
			name:            "Python 3.12 with mise 2025 - should not set",
			toolName:        "python",
			concreteVersion: "3.12.0",
			miseVersion:     "v2025.12.1",
			want:            false,
		},
		{
			name:            "Python 3.15 with mise 2025 - should set",
			toolName:        "python",
			concreteVersion: "3.15.0",
			miseVersion:     "v2025.12.1",
			want:            true,
		},
		{
			name:            "Python 3.14 with nixpkgs backend - should set",
			toolName:        "nixpkgs:python",
			concreteVersion: "3.14.0",
			miseVersion:     "v2025.12.1",
			want:            true,
		},
		{
			name:            "Node.js with mise 2025 - should not set",
			toolName:        "node",
			concreteVersion: "20.0.0",
			miseVersion:     "v2025.12.1",
			want:            false,
		},
		{
			name:            "Python 3.14.0a1 with mise 2025 - should set",
			toolName:        "python",
			concreteVersion: "3.14.0a1",
			miseVersion:     "v2025.12.1",
			want:            true,
		},
		{
			name:            "Python 4.0 with mise 2025 - should set",
			toolName:        "python",
			concreteVersion: "4.0.0",
			miseVersion:     "v2025.12.1",
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSetPythonPrecompiledFlavor(tt.toolName, tt.concreteVersion, tt.miseVersion)
			require.Equal(t, tt.want, got, "toolName=%s, version=%s, mise=%s", tt.toolName, tt.concreteVersion, tt.miseVersion)
		})
	}
}
