package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLimitEnvVarValue(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		limitInBytes int
		wantValue    string
		wantLimited  bool
	}{
		{
			name:         "no truncation if value is not exceeding the limit",
			value:        strings.Repeat("a", 5),
			limitInBytes: 5,
			wantValue:    strings.Repeat("a", 5),
			wantLimited:  false,
		},
		{
			name:         "no truncation below min limit (5 bytes)",
			value:        strings.Repeat("a", 5),
			limitInBytes: 4,
			wantValue:    strings.Repeat("a", 5),
			wantLimited:  false,
		},
		{
			name:         "truncating to min limit (5 bytes)",
			value:        strings.Repeat("a", 6),
			limitInBytes: 5,
			wantValue:    "a...a",
			wantLimited:  true,
		},
		{
			name:         "truncating to a limit",
			value:        strings.Repeat("a", 7),
			limitInBytes: 6,
			wantValue:    "a...aa",
			wantLimited:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotLimited := LimitEnvVarValue(tt.value, tt.limitInBytes)
			require.Equal(t, tt.wantValue, gotValue)
			require.Equal(t, tt.wantLimited, gotLimited)
		})
	}
}
