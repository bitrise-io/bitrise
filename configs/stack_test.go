package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEdgeStack(t *testing.T) {
	tests := []struct {
		name           string
		stackStatus    string
		stackID        string
		setStackStatus bool
		setStackID     bool
		expectedIsEdge bool
	}{
		{
			name:           "edge stack - BITRISEIO_STACK_STATUS is edge",
			stackStatus:    "edge",
			setStackStatus: true,
			expectedIsEdge: true,
		},
		{
			name:           "edge stack - BITRISEIO_STACK_ID contains edge",
			stackID:        "osx-xcode-26.2.x-edge",
			setStackID:     true,
			expectedIsEdge: true,
		},
		{
			name:           "edge stack - only BITRISEIO_STACK_STATUS set to edge",
			stackStatus:    "edge",
			stackID:        "ubuntu-noble-24.04-bitrise-2025-android",
			setStackStatus: true,
			setStackID:     true,
			expectedIsEdge: true,
		},
		{
			name:           "non-edge stack - BITRISEIO_STACK_STATUS without edge",
			stackStatus:    "stable",
			setStackStatus: true,
			expectedIsEdge: false,
		},
		{
			name:           "non-edge stack - BITRISEIO_STACK_ID without edge",
			stackID:        "osx-xcode-16.0.x",
			setStackID:     true,
			expectedIsEdge: false,
		},
		{
			name:           "non-edge stack - empty BITRISEIO_STACK_STATUS and BITRISEIO_STACK_ID",
			expectedIsEdge: false,
		},
		{
			name:           "edge stack - BITRISEIO_STACK_STATUS with edge in middle",
			stackStatus:    "stable-edge-preview",
			setStackStatus: true,
			expectedIsEdge: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BITRISEIO_STACK_STATUS", "")
			t.Setenv("BITRISEIO_STACK_ID", "")

			if tt.setStackStatus {
				t.Setenv("BITRISEIO_STACK_STATUS", tt.stackStatus)
			}
			if tt.setStackID {
				t.Setenv("BITRISEIO_STACK_ID", tt.stackID)
			}

			result := IsEdgeStack()
			require.Equal(t, tt.expectedIsEdge, result)
		})
	}
}
