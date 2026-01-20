package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEdgeStack(t *testing.T) {
	tests := []struct {
		name           string
		stackStatus    string
		setStackStatus bool
		expectedIsEdge bool
	}{
		{
			name:           "edge stack - BITRISEIO_STACK_STATUS is edge",
			stackStatus:    "edge",
			setStackStatus: true,
			expectedIsEdge: true,
		},
		{
			name:           "edge stack - BITRISEIO_STACK_STATUS with edge in middle",
			stackStatus:    "stable-edge-preview",
			setStackStatus: true,
			expectedIsEdge: true,
		},
		{
			name:           "non-edge stack - BITRISEIO_STACK_STATUS without edge",
			stackStatus:    "stable",
			setStackStatus: true,
			expectedIsEdge: false,
		},
		{
			name:           "non-edge stack - BITRISEIO_STACK_STATUS not set",
			expectedIsEdge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BITRISEIO_STACK_STATUS", "")

			if tt.setStackStatus {
				t.Setenv("BITRISEIO_STACK_STATUS", tt.stackStatus)
			}

			result := IsEdgeStack()
			require.Equal(t, tt.expectedIsEdge, result)
		})
	}
}
