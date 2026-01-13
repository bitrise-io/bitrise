package yml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateTools(t *testing.T) {
	tests := []struct {
		name        string
		config      *BitriseDataModel
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid tools with various version formats",
			config: &BitriseDataModel{
				Tools: ToolsModel{
					"python":      "3.9.0",
					"node":        "16.14.0",
					"ruby":        "2.7:latest",
					"go":          "1.19:installed",
					"java":        "11",
					"custom-tool": "1.0.0-beta.1",
				},
			},
			wantErr: false,
		},
		{
			name: "empty tools config",
			config: &BitriseDataModel{
				Tools: ToolsModel{},
			},
			wantErr: false,
		},
		{
			name: "nil tools config",
			config: &BitriseDataModel{
				Tools: nil,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid latest syntax",
			config: &BitriseDataModel{
				Tools: ToolsModel{
					"python": "3.9:latest",
					"node":   "16:latest",
					"ruby":   "2.7.4:latest",
				},
			},
			wantErr: false,
		},
		{
			name: "workflow-level tools without global tools - valid setup",
			config: &BitriseDataModel{
				Tools: nil, // No global tools
				Workflows: map[string]WorkflowModel{
					"test": {
						Tools: ToolsModel{
							"python": "3.9.0",
							"node":   "16:latest",
						},
					},
					"deploy": {
						Tools: ToolsModel{
							"ruby": "2.7:installed",
							"go":   "1.19.5",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "workflow-level tools with global tools - valid setup",
			config: &BitriseDataModel{
				Tools: ToolsModel{
					"python": "3.8.0",
				},
				Workflows: map[string]WorkflowModel{
					"test": {
						Tools: ToolsModel{
							"python": "3.9.0", // Override global python version
							"node":   "16:latest",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTools(tt.config)

			if tt.wantErr {
				require.Error(t, err, "expected an error but got none")
				if tt.expectedErr != "" {
					require.Contains(t, err.Error(), tt.expectedErr, "error message should contain expected text")
				}
			} else {
				require.NoError(t, err, "expected no error but got: %v", err)
			}
		})
	}
}
