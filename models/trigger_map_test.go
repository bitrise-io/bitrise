package models

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestTriggerMapModel_Validate(t *testing.T) {
	tests := []struct {
		name         string
		triggerMap   TriggerMapModel
		workflows    []string
		pipelines    []string
		wantErr      string
		wantWarnings []string
	}{
		{
			name: "Simple trigger items",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PushBranch: "master",
					WorkflowID: "ci",
				},
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					PullRequestTargetBranch: "master",
					DraftPullRequestEnabled: pointers.NewBoolPtr(false),
					WorkflowID:              "ci",
				},
				TriggerMapItemModel{
					Tag:        "0.9.0",
					WorkflowID: "release",
				},
			},
			workflows: []string{"ci", "release"},
		},
		{
			name: "Push trigger items",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PushBranch: "master",
					WorkflowID: "ci",
				},
				TriggerMapItemModel{
					PushBranch: "release",
					WorkflowID: "release",
				},
			},
			workflows: []string{"ci", "release"},
		},
		{
			name: "Push trigger items - duplication",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PushBranch: "master",
					WorkflowID: "ci",
				},
				TriggerMapItemModel{
					PushBranch: "master",
					WorkflowID: "release",
				},
			},
			workflows: []string{"ci", "release"},
			wantErr:   "the 2. trigger item duplicates the 1. trigger item",
		},
		{
			name: "Pull Request trigger items",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PullRequestSourceBranch: "feature*",
					WorkflowID:              "ci",
				},
				TriggerMapItemModel{
					PullRequestSourceBranch: "master",
					WorkflowID:              "release",
				},
			},
			workflows: []string{"ci", "release"},
		},
		{
			name: "Pull Request trigger items - duplicated (source branch)",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					WorkflowID:              "ci",
				},
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					WorkflowID:              "release",
				},
			},
			workflows: []string{"ci", "release"},
			wantErr:   "the 2. trigger item duplicates the 1. trigger item",
		},
		{
			name: "Pull Request trigger items - duplicated (target branch)",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PullRequestTargetBranch: "master",
					WorkflowID:              "ci",
				},
				TriggerMapItemModel{
					PullRequestTargetBranch: "master",
					WorkflowID:              "release",
				},
			},
			workflows: []string{"ci", "release"},
			wantErr:   "the 2. trigger item duplicates the 1. trigger item",
		},
		{
			name: "Pull Request trigger items - duplicated (source & target branch)",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					PullRequestTargetBranch: "master",
					WorkflowID:              "ci",
				},
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					PullRequestTargetBranch: "master",
					WorkflowID:              "release",
				},
			},
			workflows: []string{"ci", "release"},
			wantErr:   "the 2. trigger item duplicates the 1. trigger item",
		},
		{
			name: "Pull Request trigger items - different draft pr enabled",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					PullRequestTargetBranch: "master",
					DraftPullRequestEnabled: pointers.NewBoolPtr(false),
					WorkflowID:              "release",
				},
				TriggerMapItemModel{
					PullRequestSourceBranch: "develop",
					PullRequestTargetBranch: "master",
					DraftPullRequestEnabled: pointers.NewBoolPtr(true),
					WorkflowID:              "ci",
				},
			},
			workflows: []string{"ci", "release"},
		},
		{
			name: "Tag trigger items - duplicated",
			triggerMap: TriggerMapModel{
				TriggerMapItemModel{
					Tag:        "0.9.0",
					WorkflowID: "ci",
				},
				TriggerMapItemModel{
					Tag:        "0.9.0",
					WorkflowID: "release",
				},
			},
			workflows: []string{"ci", "release"},
			wantErr:   "the 2. trigger item duplicates the 1. trigger item",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := tt.triggerMap.Validate(tt.workflows, tt.pipelines)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
			require.Equal(t, tt.wantWarnings, warnings)
		})
	}
}

func TestTriggerMapModel_FirstMatchingTarget_SimplePRTriggerMap(t *testing.T) {
	triggerMap := TriggerMapModel{
		TriggerMapItemModel{
			PullRequestTargetBranch: "*",
			WorkflowID:              "ci",
		},
	}

	tests := []struct {
		name           string
		prTargetBranch string
		prReadyState   PullRequestReadyState
		wantWorkflow   string
		wantErr        string
	}{
		{
			name:           "PR with draft status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateDraft,
			wantWorkflow:   "ci",
		},
		{
			name:           "PR with ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateReadyForReview,
			wantWorkflow:   "ci",
		},
		{
			name:           "PR without status",
			prTargetBranch: "master",
			prReadyState:   "",
			wantWorkflow:   "ci",
		},
		{
			name:           "PR with converted to ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateConvertedToReadyForReview,
			wantErr:        "no matching pipeline & workflow found with trigger params: push-branch: , pr-source-branch: , pr-target-branch: master, tag: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipeline, gotWorkflow, err := triggerMap.FirstMatchingTarget("", "", tt.prTargetBranch, tt.prReadyState, "")
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
			require.Equal(t, "", gotPipeline)
			require.Equal(t, tt.wantWorkflow, gotWorkflow)
		})
	}
}

func TestTriggerMapModel_FirstMatchingTarget_DraftPRDisabledTriggerMap(t *testing.T) {
	triggerMap := TriggerMapModel{
		TriggerMapItemModel{
			PullRequestTargetBranch: "*",
			DraftPullRequestEnabled: pointers.NewBoolPtr(false),
			WorkflowID:              "ci",
		},
	}

	tests := []struct {
		name           string
		prTargetBranch string
		prReadyState   PullRequestReadyState
		wantWorkflow   string
		wantErr        string
	}{
		{
			name:           "PR with draft status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateDraft,
			wantErr:        "no matching pipeline & workflow found with trigger params: push-branch: , pr-source-branch: , pr-target-branch: master, tag: ",
		},
		{
			name:           "PR with ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateReadyForReview,
			wantWorkflow:   "ci",
		},
		{
			name:           "PR without status",
			prTargetBranch: "master",
			prReadyState:   "",
			wantWorkflow:   "ci",
		},
		{
			name:           "PR with converted to ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateConvertedToReadyForReview,
			wantWorkflow:   "ci",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipeline, gotWorkflow, err := triggerMap.FirstMatchingTarget("", "", tt.prTargetBranch, tt.prReadyState, "")
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
			require.Equal(t, "", gotPipeline)
			require.Equal(t, tt.wantWorkflow, gotWorkflow)
		})
	}
}

func TestTriggerMapModel_FirstMatchingTarget_PRTriggerMapWithDraftPRDisabledAndEnabledItems(t *testing.T) {
	triggerMap := TriggerMapModel{
		TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			DraftPullRequestEnabled: pointers.NewBoolPtr(false),
			WorkflowID:              "ci",
		},
		TriggerMapItemModel{
			PullRequestTargetBranch: "*",
			WorkflowID:              "lint",
		},
	}

	tests := []struct {
		name           string
		prTargetBranch string
		prReadyState   PullRequestReadyState
		wantWorkflow   string
		wantErr        string
	}{
		{
			name:           "PR with draft status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateDraft,
			wantWorkflow:   "lint",
		},
		{
			name:           "PR with ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateReadyForReview,
			wantWorkflow:   "ci",
		},
		{
			name:           "PR without status",
			prTargetBranch: "master",
			prReadyState:   "",
			wantWorkflow:   "ci",
		},
		{
			name:           "PR with converted to ready for review status",
			prTargetBranch: "master",
			prReadyState:   PullRequestReadyStateConvertedToReadyForReview,
			wantWorkflow:   "ci",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipeline, gotWorkflow, err := triggerMap.FirstMatchingTarget("", "", tt.prTargetBranch, tt.prReadyState, "")
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
			require.Equal(t, "", gotPipeline)
			require.Equal(t, tt.wantWorkflow, gotWorkflow)
		})
	}
}
