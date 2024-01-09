package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckDuplicatedTriggerMapItems(t *testing.T) {
	t.Log("duplicated push - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (push_branch: master)")
	}

	t.Log("duplicated pull request - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop)")

		err = checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_target_branch: master)")

		err = checkDuplicatedTriggerMapItems(TriggerMapModel{
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
		})

		require.EqualError(t, err, "duplicated trigger item found (pull_request_source_branch: develop && pull_request_target_branch: master)")
	}

	t.Log("duplicated tag - error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "ci",
			},
			TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
		})

		require.EqualError(t, err, "duplicated trigger item found (tag: 0.9.0)")
	}

	t.Log("complex trigger map - no error")
	{
		err := checkDuplicatedTriggerMapItems(TriggerMapModel{
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
		})

		require.NoError(t, err)
	}
}
