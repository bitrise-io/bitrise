package update

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsComparable(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{version: "2.45.0", want: true},
		{version: "0.0.1", want: true},
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
			require.Equal(t, tt.want, isComparable(tt.version))
		})
	}
}
