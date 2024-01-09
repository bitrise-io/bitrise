package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckDuplicatedTriggerMapItems(t *testing.T) {
	t.Log("duplicated push - error")
	{
		err := TriggerMapModel{
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "release",
			},
		}.checkDuplicatedTriggerMapItems()

		require.EqualError(t, err, "duplicated trigger item found (push_branch: master)")
	}

	t.Log("duplicated pull request - error")
	{
		err := TriggerMapModel{
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "release",
			},
		}.checkDuplicatedTriggerMapItems()

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop)")

		err = TriggerMapModel{
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "release",
			},
		}.checkDuplicatedTriggerMapItems()

		require.EqualError(t, err, "duplicated trigger item found (pull_request_target_branch: master)")

		err = TriggerMapModel{
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
		}.checkDuplicatedTriggerMapItems()

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop && pull_request_target_branch: master)")
	}

	t.Log("duplicated tag - error")
	{
		err := TriggerMapModel{
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
		}.checkDuplicatedTriggerMapItems()

		require.EqualError(t, err, "duplicated trigger item found (tag: 0.9.0)")
	}

	t.Log("complex trigger map - no error")
	{
		err := TriggerMapModel{
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
		}.checkDuplicatedTriggerMapItems()

		require.NoError(t, err)
	}
}

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
			name: "duplicated push - error",
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
			name: "duplicated pull request - error",
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
			wantErr:   "duplicated trigger item found (push_branch: master)",
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
