package versionresolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nodeVersions = []string{
	"14.0.0", "14.21.3",
	"16.0.0", "16.20.2",
	"17.0.0",
	"18.0.0", "18.1.0", "18.19.0", "18.20.4",
	"19.0.0",
	"20.0.0", "20.10.0", "20.18.1",
	"21.0.0",
	"22.0.0", "22.5.0",
}

func TestResolveConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		want       string
	}{
		{
			name:       "caret major",
			constraint: "^18.0.0",
			want:       "18.20.4",
		},
		{
			name:       "caret minor",
			constraint: "^18.1.0",
			want:       "18.20.4",
		},
		{
			name:       "tilde",
			constraint: "~18.1.0",
			want:       "18.1.0",
		},
		{
			name:       "gte",
			constraint: ">=18",
			want:       "22.5.0",
		},
		{
			name:       "gte with lt",
			constraint: ">=18, <20",
			want:       "19.0.0",
		},
		{
			name:       "space-separated AND",
			constraint: ">=16 <20",
			want:       "19.0.0",
		},
		{
			name:       "wildcard x",
			constraint: "18.x",
			want:       "18.20.4",
		},
		{
			name:       "wildcard star",
			constraint: "18.*",
			want:       "18.20.4",
		},
		{
			name:       "bare major",
			constraint: "18",
			want:       "18.20.4",
		},
		{
			name:       "exact version",
			constraint: "18.0.0",
			want:       "18.0.0",
		},
		{
			name:       "OR branches",
			constraint: "14.x || 16.x || 18.x",
			want:       "18.20.4",
		},
		{
			name:       "hyphen range",
			constraint: "1.2.3 - 18.1.0",
			want:       "18.1.0",
		},
		{
			name:       "star matches all",
			constraint: "*",
			want:       "22.5.0",
		},
		{
			name:       "typical LTS range",
			constraint: ">=18.0.0 <23.0.0",
			want:       "22.5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveConstraint(tt.constraint, nodeVersions)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveConstraint_NoMatch(t *testing.T) {
	_, err := ResolveConstraint(">=99", nodeVersions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no version matching")
}

func TestResolveConstraint_InvalidConstraint(t *testing.T) {
	_, err := ResolveConstraint("not-a-constraint!!!", nodeVersions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid semver constraint")
}

func TestResolveConstraint_Empty(t *testing.T) {
	_, err := ResolveConstraint("", nodeVersions)
	assert.Error(t, err)
}

func TestResolveConstraint_SkipsPrerelease(t *testing.T) {
	versions := []string{"20.0.0-rc.1", "20.0.0", "19.0.0"}
	got, err := ResolveConstraint(">=20.0.0", versions)
	require.NoError(t, err)
	assert.Equal(t, "20.0.0", got)
}

func TestResolveConstraint_Caret0x(t *testing.T) {
	versions := []string{"0.2.0", "0.2.3", "0.2.9", "0.3.0", "1.0.0"}
	got, err := ResolveConstraint("^0.2.3", versions)
	require.NoError(t, err)
	assert.Equal(t, "0.2.9", got)
}
