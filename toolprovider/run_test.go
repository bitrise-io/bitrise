package toolprovider

import (
	"fmt"
	"testing"
	"time"

	clianalytics "github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/stepman/activator"
	"github.com/stretchr/testify/require"
    "github.com/stretchr/testify/assert"
)

type trackingCall struct {
	provider     string
	request      provider.ToolRequest
	result       provider.ToolInstallResult
	isSuccessful bool
}

type capturingTracker struct {
	toolSetupCalls []trackingCall
}

func (c *capturingTracker) SendWorkflowStarted(analytics.Properties, string, string) {}
func (c *capturingTracker) SendWorkflowFinished(analytics.Properties, bool)          {}
func (c *capturingTracker) SendStepStartedEvent(analytics.Properties, clianalytics.StepInfo, time.Duration, map[string]interface{}, map[string]string) {
}
func (c *capturingTracker) SendStepFinishedEvent(analytics.Properties, clianalytics.StepResult) {}
func (c *capturingTracker) SendCLIWarning(string)                                               {}
func (c *capturingTracker) SendCommandInfo(string, string, []string)                            {}
func (c *capturingTracker) SendToolSetupEvent(providerID string, request provider.ToolRequest, result provider.ToolInstallResult, isSuccessful bool, setupTime time.Duration) {
	c.toolSetupCalls = append(c.toolSetupCalls, trackingCall{
		provider:     providerID,
		request:      request,
		result:       result,
		isSuccessful: isSuccessful,
	})
}
func (c *capturingTracker) SendStepActivationEvent(activator.ActivationType, string, bool, time.Duration, bool) {
}
func (c *capturingTracker) Wait()            {}
func (c *capturingTracker) IsTracking() bool { return true }

type fakeToolProvider struct {
	installResult provider.ToolInstallResult
	installErr    error
	activation    provider.EnvironmentActivation
	activationErr error
}

func (f fakeToolProvider) ID() string { return "fake" }

func (f fakeToolProvider) Bootstrap() error { return nil }

func (f fakeToolProvider) InstallTool(provider.ToolRequest) (provider.ToolInstallResult, error) {
	return f.installResult, f.installErr
}

func (f fakeToolProvider) ActivateEnv(provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	return f.activation, f.activationErr
}

func (f fakeToolProvider) ListReleasedVersions(provider.ToolID) ([]string, error) {
	return nil, nil
}

func TestInstallResolvedToolsReportsToolInstallError(t *testing.T) {
	tracker := &capturingTracker{}
	request := provider.ToolRequest{
		ToolName:        "nodejs",
		UnparsedVersion: "999.0.0",
	}

	_, err := installResolvedTools([]provider.ToolRequest{request}, "mise", fakeToolProvider{
		installErr: provider.ToolInstallError{
			ToolName:         request.ToolName,
			RequestedVersion: request.UnparsedVersion,
			Cause:            "no matching version",
		},
	}, tracker, true, time.Now())

	require.EqualError(t, err, "see error details above")
	require.Len(t, tracker.toolSetupCalls, 1)
	require.Equal(t, trackingCall{
		provider: "mise",
		request: provider.ToolRequest{
			ToolName:        "nodejs",
			UnparsedVersion: "999.0.0",
		},
		result:       provider.ToolInstallResult{},
		isSuccessful: false,
	}, tracker.toolSetupCalls[0])
}

func TestInstallResolvedToolsReportsActivationFailure(t *testing.T) {
	tracker := &capturingTracker{}
	request := provider.ToolRequest{
		ToolName:        "golang",
		UnparsedVersion: "1.22.0",
	}
	result := provider.ToolInstallResult{
		ToolName:           "golang",
		ConcreteVersion:    "1.22.0",
		IsAlreadyInstalled: true,
	}

	_, err := installResolvedTools([]provider.ToolRequest{request}, "asdf", fakeToolProvider{
		installResult: result,
		activationErr: fmt.Errorf("activation failed"),
	}, tracker, true, time.Now())

	require.EqualError(t, err, "activate golang: activation failed")
	require.Len(t, tracker.toolSetupCalls, 1)
	require.Equal(t, trackingCall{
		provider:     "asdf",
		request:      request,
		result:       result,
		isSuccessful: false,
	}, tracker.toolSetupCalls[0])
}

func TestInstallResolvedToolsReportsSuccessAfterActivation(t *testing.T) {
	tracker := &capturingTracker{}
	request := provider.ToolRequest{
		ToolName:        "ruby",
		UnparsedVersion: "3.3.0",
	}
	result := provider.ToolInstallResult{
		ToolName:        "ruby",
		ConcreteVersion: "3.3.0",
	}
	activation := provider.EnvironmentActivation{
		ContributedEnvVars: map[string]string{"GEM_HOME": "/tmp/gems"},
	}

	activations, err := installResolvedTools([]provider.ToolRequest{request}, "mise", fakeToolProvider{
		installResult: result,
		activation:    activation,
	}, tracker, true, time.Now())

	require.NoError(t, err)
	require.Equal(t, []provider.EnvironmentActivation{activation}, activations)
	require.Len(t, tracker.toolSetupCalls, 1)
	require.Equal(t, trackingCall{
		provider:     "mise",
		request:      request,
		result:       result,
		isSuccessful: true,
	}, tracker.toolSetupCalls[0])
}

func TestFindGitHubTokenEnv(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedName  string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "empty map",
			envs:          map[string]string{},
			expectedFound: false,
		},
		{
			name:          "no known tokens",
			envs:          map[string]string{"SOME_OTHER_VAR": "value"},
			expectedFound: false,
		},
		{
			name:          "GITHUB_TOKEN present",
			envs:          map[string]string{"GITHUB_TOKEN": "ghp_abc123"},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "ghp_abc123",
			expectedFound: true,
		},
		{
			name:          "MISE_GITHUB_TOKEN present",
			envs:          map[string]string{"MISE_GITHUB_TOKEN": "ghp_xyz"},
			expectedName:  "MISE_GITHUB_TOKEN",
			expectedValue: "ghp_xyz",
			expectedFound: true,
		},
		{
			name: "GITHUB_TOKEN takes priority over MISE_GITHUB_TOKEN",
			envs: map[string]string{
				"GITHUB_TOKEN":      "github_value",
				"MISE_GITHUB_TOKEN": "mise_value",
			},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "github_value",
			expectedFound: true,
		},
		{
			name: "GITHUB_TOKEN takes priority over GITHUB_API_TOKEN",
			envs: map[string]string{
				"GITHUB_TOKEN":     "github_value",
				"GITHUB_API_TOKEN": "api_value",
			},
			expectedName:  "GITHUB_TOKEN",
			expectedValue: "github_value",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, found := findGitHubTokenEnv(tt.envs)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
