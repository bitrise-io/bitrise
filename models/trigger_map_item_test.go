package models

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestTriggerMapItemModel_MatchWithParams_CodePushParams(t *testing.T) {
	t.Log("The following patterns are all matches")
	{
		for aPattern, aPushBranch := range map[string]string{
			"feature":   "feature",
			"feature/*": "feature/login",
			"feature**": "feature",
			"*feature":  "feature",
			"**feature": "feature",
			"*":         "feature",
		} {
			pushBranch := aPushBranch
			prSourceBranch := ""
			prTargetBranch := ""
			tag := ""

			item := TriggerMapItemModel{
				PushBranch: aPattern,
				WorkflowID: "primary",
			}
			match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
			require.NoError(t, err)
			require.Equal(t, true, match, "(pattern: %s) (branch: %s)", aPattern, aPushBranch)
		}
	}

	t.Log("code-push against code-push type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PushBranch: "deploy",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			WorkflowID:              "test",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("code-push against pr type item - NOT MATCH")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		item := TriggerMapItemModel{
			PullRequestSourceBranch: "develop",
			PullRequestTargetBranch: "master",
			WorkflowID:              "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}
}

func TestTriggerMapItemModel_MatchWithParams_PRParams(t *testing.T) {
	tests := []struct {
		name           string
		triggerMapItem TriggerMapItemModel
		pushBranch     string
		prSourceBranch string
		prTargetBranch string
		isDraftPR      bool
		tag            string
		want           bool
		wantErr        string
	}{
		// Match tests
		{
			name: "pr against pr type item - MATCH (source branch)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           true,
		},
		{
			name: "pr against pr type item - MATCH (target branch)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           true,
		},
		{
			name: "pr against pr type item - MATCH (target and source branch)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           true,
		},
		{
			name: "pr against pr type item (simple glob source branch) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "*",
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           true,
		},
		{
			name: "pr against pr type item (glob target branch) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "deploy_*",
				WorkflowID:              "primary",
			},
			prTargetBranch: "deploy_1_0_0",
			want:           true,
		},
		{
			name: "pr against pr type item (complex glob source branch) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature/*",
				PullRequestTargetBranch: "develop",
				WorkflowID:              "test",
			},
			prSourceBranch: "feature/login",
			prTargetBranch: "develop",
			want:           true,
		},
		{
			name: "draft pr against pr type item (draft pr explicitly enabled) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				DraftPullRequestEnabled: pointers.NewBoolPtr(true),
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			isDraftPR:      true,
			want:           true,
		},
		{
			name: "draft pr against pr type item (draft pr enabled by default) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			isDraftPR:      true,
			want:           true,
		},
		// No match tests
		{
			name: "pr against pr type item - NOT MATCH (target branch mismatch)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "deploy",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           false,
		},
		{
			name: "pr against pr type item - NOT MATCH (source branch mismatch)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature/*",
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           false,
		},
		{
			name: "pr against push type item - NOT MATCH",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			want:           false,
		},
		{
			name: "draft pr against pr type item (draft pr explicitly disabled) - MATCH",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				DraftPullRequestEnabled: pointers.NewBoolPtr(false),
				WorkflowID:              "primary",
			},
			prSourceBranch: "develop",
			prTargetBranch: "master",
			isDraftPR:      true,
			want:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.triggerMapItem.MatchWithParams(tt.pushBranch, tt.prSourceBranch, tt.prTargetBranch, PullRequestReadyStateDraft, tt.tag)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTriggerMapItemModel_MatchWithParams_TagParams(t *testing.T) {
	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		item := TriggerMapItemModel{
			Tag:        "0.9.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		item := TriggerMapItemModel{
			Tag:        "0.9.0",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			Tag:        "0.9.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, true, match)
	}

	t.Log("tag against tag type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			Tag:        "1.*",
			WorkflowID: "deploy",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}

	t.Log("tag against push type item - NOT MATCH")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0-pre"

		item := TriggerMapItemModel{
			PushBranch: "master",
			WorkflowID: "primary",
		}
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, PullRequestReadyStateReadyForReview, tag)
		require.NoError(t, err)
		require.Equal(t, false, match)
	}
}

func TestTriggerMapItemModel_String(t *testing.T) {
	tests := []struct {
		name                string
		triggerMapItem      TriggerMapItemModel
		want                string
		wantWithPrintTarget string
	}{
		{
			name: "triggering pipeline",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				PipelineID: "pipeline-1",
			},
			want:                "push_branch: master",
			wantWithPrintTarget: "push_branch: master -> pipeline: pipeline-1",
		},
		{
			name: "push event",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			want:                "push_branch: master",
			wantWithPrintTarget: "push_branch: master -> workflow: ci",
		},
		{
			name: "pull request event - pr source branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "ci",
			},
			want:                "pull_request_source_branch: develop && draft_pull_request_enabled: true",
			wantWithPrintTarget: "pull_request_source_branch: develop && draft_pull_request_enabled: true -> workflow: ci",
		},
		{
			name: "pull request event - pr target branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			want:                "pull_request_target_branch: master && draft_pull_request_enabled: true",
			wantWithPrintTarget: "pull_request_target_branch: master && draft_pull_request_enabled: true -> workflow: ci",
		},
		{
			name: "pull request event - pr target and source branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			want:                "pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: true",
			wantWithPrintTarget: "pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: true -> workflow: ci",
		},
		{
			name: "pull request event - pr target and source branch and disable draft prs",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				DraftPullRequestEnabled: pointers.NewBoolPtr(false),
				WorkflowID:              "ci",
			},
			want:                "pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: false",
			wantWithPrintTarget: "pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: false -> workflow: ci",
		},
		{
			name: "tag event",
			triggerMapItem: TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
			want:                "tag: 0.9.0",
			wantWithPrintTarget: "tag: 0.9.0 -> workflow: release",
		},
		{
			name: "deprecated type - pr disabled",
			triggerMapItem: TriggerMapItemModel{
				Pattern:              "master",
				IsPullRequestAllowed: false,
				WorkflowID:           "ci",
			},
			want:                "pattern: master && is_pull_request_allowed: false",
			wantWithPrintTarget: "pattern: master && is_pull_request_allowed: false -> workflow: ci",
		},
		{
			name: "deprecated type - pr enabled",
			triggerMapItem: TriggerMapItemModel{
				Pattern:              "master",
				IsPullRequestAllowed: true,
				WorkflowID:           "ci",
			},
			want:                "pattern: master && is_pull_request_allowed: true",
			wantWithPrintTarget: "pattern: master && is_pull_request_allowed: true -> workflow: ci",
		},
		{
			name: "mixed",
			triggerMapItem: TriggerMapItemModel{
				PushBranch:              "master",
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				Tag:                     "0.9.0",
				Pattern:                 "*",
				IsPullRequestAllowed:    true,
				WorkflowID:              "ci",
			},
			want:                "push_branch: master pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: true tag: 0.9.0 pattern: * && is_pull_request_allowed: true",
			wantWithPrintTarget: "push_branch: master pull_request_source_branch: develop && pull_request_target_branch: master && draft_pull_request_enabled: true tag: 0.9.0 pattern: * && is_pull_request_allowed: true -> workflow: ci",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.triggerMapItem.conditionsString())
			require.Equal(t, tt.wantWithPrintTarget, tt.triggerMapItem.conditionsString())
		})
	}
}

func TestTriggerMapItemModel_Validate(t *testing.T) {
	tests := []struct {
		name           string
		triggerMapItem TriggerMapItemModel
		workflows      []string
		pipelines      []string
		wantWarns      []string
		wantErr        string
	}{
		{
			name: "it validates deprecated trigger item with triggered pipeline",
			triggerMapItem: TriggerMapItemModel{
				Pattern:    "*",
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it validates deprecated trigger item with triggered workflow",
			triggerMapItem: TriggerMapItemModel{
				Pattern:    "*",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "it fails for invalid deprecated trigger item - pipeline & workflow both defined",
			triggerMapItem: TriggerMapItemModel{
				Pattern:    "*",
				PipelineID: "pipeline-1",
				WorkflowID: "workflow-1",
			},
			workflows: []string{"pipeline-1", "workflow-1"},
			wantErr:   "both pipeline and workflow are defined as trigger target: pattern: * && is_pull_request_allowed: false",
		},
		{
			name: "it fails for invalid deprecated trigger item - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				Pattern: "*",
			},
			wantErr: "no pipeline nor workflow is defined as a trigger target: pattern: * && is_pull_request_allowed: false",
		},
		{
			name: "it fails for invalid deprecated trigger item - missing pattern",
			triggerMapItem: TriggerMapItemModel{
				Pattern:    "",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger map item ( -> workflow: primary) validate failed, error: failed to determin trigger event from params: push-branch: , pr-source-branch: , pr-target-branch: , tag: ",
		},
		{
			name: "it validates code-push trigger item with triggered pipeline",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "*",
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it validates code-push trigger item with triggered workflow",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "*",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "it fails for invalid code-push trigger item - missing push-branch",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger map item ( -> workflow: primary) validate failed, error: failed to determin trigger event from params: push-branch: , pr-source-branch: , pr-target-branch: , tag: ",
		},
		{
			name: "it fails for invalid code-push trigger item - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "*",
			},
			wantErr: "no pipeline nor workflow is defined as a trigger target: push_branch: *",
		},
		{
			name: "it validates pull-request trigger item (with source branch) with triggered pipeline",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature/",
				PipelineID:              "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it validates pull-request trigger item (with source branch) with triggered workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature/",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "it validates pull-request trigger item (with target branch) with triggered pipeline",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				PipelineID:              "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it validates pull-request trigger item (with target branch) with triggered workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "it fails for invalid pull-request trigger item (target branch set) - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "*",
			},
			wantErr: "no pipeline nor workflow is defined as a trigger target: pull_request_target_branch: * && draft_pull_request_enabled: true",
		},
		{
			name: "it fails for invalid pull-request trigger item (target and source branch set) - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature*",
				PullRequestTargetBranch: "master",
			},
			wantErr: "no pipeline nor workflow is defined as a trigger target: pull_request_source_branch: feature* && pull_request_target_branch: master && draft_pull_request_enabled: true",
		},
		{
			name: "it fails for mixed (mixed types) trigger item",
			triggerMapItem: TriggerMapItemModel{
				PushBranch:              "master",
				PullRequestSourceBranch: "feature/*",
				PullRequestTargetBranch: "",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger map item (push_branch: master pull_request_source_branch: feature/* && draft_pull_request_enabled: true -> workflow: primary) validate failed, error: push_branch (master) selects code-push trigger event, but pull_request_source_branch (feature/*) also provided",
		},
		{
			name: "it fails for mixed (mixed new and legacy properties) trigger item",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				Pattern:    "*",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "deprecated trigger item (pattern defined), mixed with trigger params (push_branch: master, pull_request_source_branch: , pull_request_target_branch: , tag: )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns, err := tt.triggerMapItem.Validate(0, tt.workflows, tt.pipelines)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantWarns, warns)
		})
	}
}

func TestTriggerEventType(t *testing.T) {
	t.Log("it determines trigger event type")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypeCodePush, event)
	}

	t.Log("it determines trigger event type")
	{
		pushBranch := ""
		prSourceBranch := "develop"
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypePullRequest, event)
	}

	t.Log("it determines trigger event type")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := "master"
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypePullRequest, event)
	}

	t.Log("it determines trigger event type")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.NoError(t, err)
		require.Equal(t, TriggerEventTypeTag, event)
	}

	t.Log("it fails without inputs")
	{
		pushBranch := ""
		prSourceBranch := ""
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := "develop"
		prTargetBranch := ""
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := "master"
		tag := ""

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}

	t.Log("it fails if event type not clear")
	{
		pushBranch := "master"
		prSourceBranch := ""
		prTargetBranch := ""
		tag := "0.9.0"

		event, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
		require.Error(t, err)
		require.Equal(t, TriggerEventTypeUnknown, event)
	}
}
