package toolprovider

import (
	"bytes"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

// setupTestLogger configures a logger that captures output to a buffer for testing
func setupTestLogger() *bytes.Buffer {
	var buf bytes.Buffer
	log.InitGlobalLogger(log.LoggerOpts{
		LoggerType:      log.ConsoleLogger,
		Producer:        log.BitriseCLI,
		DebugLogEnabled: true,
		Writer:          &buf,
	})
	return &buf
}

func TestPrintToolRequests(t *testing.T) {
	tests := []struct {
		name           string
		toolRequests   []provider.ToolRequest
		expectedOutput string
	}{
		{
			name: "Single tool with strict resolution",
			toolRequests: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
			},
			expectedOutput: "\n\x1b[34;1mTool setup\n\x1b[0mPlan:\n• \x1b[35;1mgolang\x1b[0m \x1b[36;1m1.20.3\x1b[0m \n\nInstalling missing tools\n",
		},
		{
			name: "Multiple tools with different resolution strategies",
			toolRequests: []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "1.20.3",
					ResolutionStrategy: provider.ResolutionStrategyStrict,
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "20",
					ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
				},
				{
					ToolName:           "ruby",
					UnparsedVersion:    "3.2",
					ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
				},
			},
			expectedOutput: "\n\x1b[34;1mTool setup\n\x1b[0mPlan:\n• \x1b[35;1mgolang\x1b[0m \x1b[36;1m1.20.3\x1b[0m \n• \x1b[35;1mnodejs\x1b[0m \x1b[36;1m20\x1b[0m (resolve to latest installed)\n• \x1b[35;1mruby\x1b[0m \x1b[36;1m3.2\x1b[0m (resolve to latest released)\n\nInstalling missing tools\n",
		},
		{
			name: "Special version strings",
			toolRequests:   []provider.ToolRequest{
				{
					ToolName:           "golang",
					UnparsedVersion:    "latest",
				},
				{
					ToolName:           "nodejs",
					UnparsedVersion:    "installed",
				},
				{
					ToolName:           "python",
					UnparsedVersion:    "",
				},
			},
			expectedOutput: "\n\x1b[34;1mTool setup\n\x1b[0mPlan:\n• \x1b[35;1mgolang\x1b[0m \x1b[36;1mlatest\x1b[0m \n• \x1b[35;1mnodejs\x1b[0m \x1b[36;1minstalled\x1b[0m \n• \x1b[35;1mpython\x1b[0m \x1b[36;1m<unset version>\x1b[0m \n\nInstalling missing tools\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := setupTestLogger()
			printToolRequests(tt.toolRequests)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

func TestPrintInstallStart(t *testing.T) {
	tests := []struct {
		name           string
		toolRequest    provider.ToolRequest
		expectedOutput string
	}{
		{
			name: "Basic tool request",
			toolRequest: provider.ToolRequest{
				ToolName:        "golang",
				UnparsedVersion: "1.20.3",
			},
			expectedOutput: "• \x1b[35;1mgolang\x1b[0m \x1b[36;1m1.20.3\x1b[0m...\n",
		},
		{
			name: "Tool request with empty version",
			toolRequest: provider.ToolRequest{
				ToolName:        "python",
				UnparsedVersion: "", // provider-specific handling for empty versions
			},
			expectedOutput: "• \x1b[35;1mpython\x1b[0m \x1b[36;1m<unset version>\x1b[0m...\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := setupTestLogger()
			printInstallStart(tt.toolRequest)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

func TestPrintInstallResult(t *testing.T) {
	duration := time.Millisecond * 1234
	tests := []struct {
		name           string
		toolRequest    provider.ToolRequest
		result         provider.ToolInstallResult
		expectedOutput string
	}{
		{
			name: "Already installed with same version",
			toolRequest: provider.ToolRequest{
				ToolName:        "golang",
				UnparsedVersion: "1.20.3",
			},
			result: provider.ToolInstallResult{
				ToolName:           "golang",
				IsAlreadyInstalled: true,
				ConcreteVersion:    "1.20.3",
			},
			expectedOutput: "\x1b[32;1m✓\x1b[0m already installed (took 1.234s)\n\n",
		},
		{
			name: "Newly installed with same version",
			toolRequest: provider.ToolRequest{
				ToolName:        "nodejs",
				UnparsedVersion: "20.0.0",
			},
			result: provider.ToolInstallResult{
				ToolName:           "nodejs",
				IsAlreadyInstalled: false,
				ConcreteVersion:    "20.0.0",
			},
			expectedOutput: "\x1b[32;1m✓\x1b[0m installed (took 1.234s)\n\n",
		},
		{
			name: "Installed with different concrete version",
			toolRequest: provider.ToolRequest{
				ToolName:        "ruby",
				UnparsedVersion: "3.2",
			},
			result: provider.ToolInstallResult{
				ToolName:           "ruby",
				IsAlreadyInstalled: false,
				ConcreteVersion:    "3.2.1",
			},
			expectedOutput: "\x1b[32;1m✓\x1b[0m installed → \x1b[36;1m3.2.1\x1b[0m (took 1.234s)\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := setupTestLogger()

			printInstallResult(tt.toolRequest, tt.result, duration)

			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

func TestPrintInstallError(t *testing.T) {
	tests := []struct {
		name           string
		err            provider.ToolInstallError
		expectedOutput string
	}{
		{
			name: "Basic error",
			err: provider.ToolInstallError{
				ToolName:         "golang",
				RequestedVersion: "1.20.3",
			},
			expectedOutput: "\x1b[31;1m⨯ install golang 1.20.3\x1b[0m\n\n",
		},
		{
			name: "Error with cause",
			err: provider.ToolInstallError{
				ToolName:         "nodejs",
				RequestedVersion: "20.0.0",
				Cause:            "Network timeout",
			},
			expectedOutput: "\x1b[31;1m⨯ install nodejs 20.0.0\x1b[0m\n  • \x1b[35;1mCause:\x1b[0m Network timeout\n\n",
		},
		{
			name: "Error with recommendation",
			err: provider.ToolInstallError{
				ToolName:         "ruby",
				RequestedVersion: "3.2.1",
				Recommendation:   "Try running with --verbose flag",
			},
			expectedOutput: "\x1b[31;1m⨯ install ruby 3.2.1\x1b[0m\n  • \x1b[35;1mRecommendation:\x1b[0m Try running with --verbose flag\n\n",
		},
		{
			name: "Error with raw output",
			err: provider.ToolInstallError{
				ToolName:         "python",
				RequestedVersion: "3.11",
				RawOutput:        "Build failed with exit code 1",
			},
			expectedOutput: "\x1b[31;1m⨯ install python 3.11\x1b[0m\n  • \x1b[35;1mRaw output:\x1b[0m Build failed with exit code 1\n\n",
		},
		{
			name: "Error with all optional fields",
			err: provider.ToolInstallError{
				ToolName:         "flutter",
				RequestedVersion: "3.16.0",
				Cause:            "Missing dependencies",
				Recommendation:   "Install required system packages",
				RawOutput:        "Error: dart command not found",
			},
			expectedOutput: "\x1b[31;1m⨯ install flutter 3.16.0\x1b[0m\n  • \x1b[35;1mCause:\x1b[0m Missing dependencies\n  • \x1b[35;1mRecommendation:\x1b[0m Install required system packages\n  • \x1b[35;1mRaw output:\x1b[0m Error: dart command not found\n\n",
		},
		{
			name: "Error with empty optional fields excludes empty field labels",
			err: provider.ToolInstallError{
				ToolName:         "test-tool",
				RequestedVersion: "1.0.0",
				Cause:            "",
				Recommendation:   "",
				RawOutput:        "",
			},
			expectedOutput: "\x1b[31;1m⨯ install test-tool 1.0.0\x1b[0m\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := setupTestLogger()

			printInstallError(tt.err)

			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}
