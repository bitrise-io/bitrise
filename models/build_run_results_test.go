package models

import (
	"testing"

	"github.com/bitrise-io/bitrise/v3/exitcode"
	"github.com/stretchr/testify/require"
)

func TestBuildRunResultsModel_FailedSkippableStepDoesNotFailBuild(t *testing.T) {
	buildRes := BuildRunResultsModel{
		SuccessSteps:         []StepRunResultsModel{{Status: StepRunStatusCodeSuccess, Idx: 0}},
		FailedSkippableSteps: []StepRunResultsModel{{Status: StepRunStatusCodeFailedSkippable, Idx: 1}},
	}

	require.True(t, buildRes.HasFailedSkippableSteps())
	require.False(t, buildRes.IsBuildFailed())
	require.Equal(t, exitcode.CLISuccess, buildRes.ExitCode())
}
