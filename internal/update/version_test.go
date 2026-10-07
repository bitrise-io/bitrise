package update

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		wantOK  bool
	}{
		{version: "2.45.0", wantOK: true},
		{version: "0.0.1", wantOK: true},
		{version: "dev"},
		{version: "2.45.1-next"},
		{version: "2.45.0-rc1"},
		{version: "2.45.0+build.7"},
		{version: "v2.45.0"},
		{version: "2.45"},
		{version: ""},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := parseVersion(tt.version)
			if !tt.wantOK {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.version, v.String())
		})
	}
}
