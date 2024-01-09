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
			wantErr:   "duplicated trigger item found (push_branch: master)",
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
			wantErr:   "duplicated trigger item found (pull_request_source_branch: develop && draft_pull_request_enabled: true)",
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
			wantErr:   "duplicated trigger item found (pull_request_target_branch: master && draft_pull_request_enabled: true)",
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
			wantErr:   "duplicated trigger item found (pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: true)",
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
			wantErr:   "duplicated trigger item found (tag: 0.9.0)",
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
