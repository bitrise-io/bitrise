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
		name           string
		triggerMapItem TriggerMapItemModel
		want           string
	}{
		{
			name: "triggering pipeline",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				PipelineID: "pipeline-1",
			},
			want: "push_branch: master",
		},
		{
			name: "push event",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				WorkflowID: "ci",
			},
			want: "push_branch: master",
		},
		{
			name: "push event - type only",
			triggerMapItem: TriggerMapItemModel{
				Type:       "push",
				WorkflowID: "ci",
			},
			want: "type: push",
		},
		{
			name: "pull request event - pr source branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				WorkflowID:              "ci",
			},
			want: "pull_request_source_branch: develop",
		},
		{
			name: "pull request event - pr target branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			want: "pull_request_target_branch: master",
		},
		{
			name: "pull request event - pr target and source branch",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				WorkflowID:              "ci",
			},
			want: "pull_request_source_branch: develop & pull_request_target_branch: master",
		},
		{
			name: "pull request event - pr target and source branch and disable draft prs",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "develop",
				PullRequestTargetBranch: "master",
				DraftPullRequestEnabled: pointers.NewBoolPtr(false),
				WorkflowID:              "ci",
			},
			want: "pull_request_source_branch: develop & pull_request_target_branch: master & draft_pull_request_enabled: false",
		},
		{
			name: "tag event",
			triggerMapItem: TriggerMapItemModel{
				Tag:        "0.9.0",
				WorkflowID: "release",
			},
			want: "tag: 0.9.0",
		},
		{
			name: "deprecated type - pr disabled",
			triggerMapItem: TriggerMapItemModel{
				Pattern:              "master",
				IsPullRequestAllowed: false,
				WorkflowID:           "ci",
			},
			want: "pattern: master",
		},
		{
			name: "deprecated type - pr enabled",
			triggerMapItem: TriggerMapItemModel{
				Pattern:              "master",
				IsPullRequestAllowed: true,
				WorkflowID:           "ci",
			},
			want: "pattern: master & is_pull_request_allowed: true",
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
			want: "push_branch: master & tag: 0.9.0 & pull_request_source_branch: develop & pull_request_target_branch: master & pattern: * & is_pull_request_allowed: true",
		},
		{
			name: "pr event - all conditions",
			triggerMapItem: TriggerMapItemModel{
				Type:               "pull_request",
				WorkflowID:         "ci",
				PullRequestComment: "my comment",
				PullRequestLabel:   "my label",
				CommitMessage:      "my commit",
				ChangedFiles:       "my file",
			},
			want: "commit_message: my commit & changed_files: my file & pull_request_label: my label & pull_request_comment: my comment",
		},
		{
			name: "push event - all conditions",
			triggerMapItem: TriggerMapItemModel{
				Type:          "push",
				WorkflowID:    "ci",
				CommitMessage: "my commit",
				ChangedFiles:  "my file",
			},
			want: "commit_message: my commit & changed_files: my file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.triggerMapItem.conditionsString())
		})
	}
}

func TestTriggerMapItemModel_Validate_LegacyItem(t *testing.T) {
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
			wantErr:   "trigger item #1: both pipeline and workflow are defined as trigger target",
		},
		{
			name: "it fails for invalid deprecated trigger item - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				Pattern: "*",
			},
			wantErr: "trigger item #1: no pipeline nor workflow is defined as trigger target",
		},
		{
			name: "it fails for invalid deprecated trigger item - missing pattern",
			triggerMapItem: TriggerMapItemModel{
				Pattern:    "",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "it fails for mixed (mixed new and legacy properties) trigger item",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "master",
				Pattern:    "*",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both pattern and push_branch defined",
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

func TestTriggerMapItemModel_Validate_CodePushItem(t *testing.T) {
	tests := []struct {
		name           string
		triggerMapItem TriggerMapItemModel
		workflows      []string
		pipelines      []string
		wantWarns      []string
		wantErr        string
	}{
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
			name: "type is required, when no condition defined",
			triggerMapItem: TriggerMapItemModel{
				Type:       CodePushType,
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "type is required, when no push_branch defined",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no push_branch defined (commit_message)",
			triggerMapItem: TriggerMapItemModel{
				CommitMessage: "CI",
				WorkflowID:    "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no push_branch defined (changed_files)",
			triggerMapItem: TriggerMapItemModel{
				ChangedFiles: "./ios/",
				WorkflowID:   "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "it fails for invalid code-push trigger item - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: "*",
			},
			wantErr: "trigger item #1: no pipeline nor workflow is defined as trigger target",
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
			wantErr:   "trigger item #1: both push_branch and pull_request_source_branch defined",
		},
		{
			name: "push_branch can be a regex",
			triggerMapItem: TriggerMapItemModel{
				PushBranch: map[interface{}]interface{}{
					"regex": "feature-.*",
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "commit_message can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: CodePushType,
				CommitMessage: map[string]interface{}{
					"regex": `^\[CI]\.*`,
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "changed_files can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: CodePushType,
				ChangedFiles: map[string]string{
					"regex": `^\/ios/.*`,
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "condition value can be a hash with a regex key",
			triggerMapItem: TriggerMapItemModel{
				Type: CodePushType,
				ChangedFiles: map[string]string{
					"glob": `^\/ios/.*`,
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: 'regex' key is expected for regex condition in changed_files field",
		},
		{
			name: "condition value can be a hash with a single key",
			triggerMapItem: TriggerMapItemModel{
				Type: CodePushType,
				ChangedFiles: map[string]string{
					"glob":  `^\/ios/*`,
					"regex": `^\/ios/.*`,
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: single 'regex' key is expected for regex condition in changed_files field",
		},
		{
			name: "invalid type",
			triggerMapItem: TriggerMapItemModel{
				Type:       "code-push",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: invalid type (code-push) defined, valid types are: push, pull_request and tag",
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

func TestTriggerMapItemModel_Validate_TagPushItem(t *testing.T) {
	tests := []struct {
		name           string
		triggerMapItem TriggerMapItemModel
		workflows      []string
		pipelines      []string
		wantWarns      []string
		wantErr        string
	}{
		{
			name: "it validates tag trigger item with triggered pipeline",
			triggerMapItem: TriggerMapItemModel{
				Tag:        "*",
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it validates tag trigger item with triggered workflow",
			triggerMapItem: TriggerMapItemModel{
				Tag:        "*",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "type is required, when no condition defined",
			triggerMapItem: TriggerMapItemModel{
				Type:       TagPushType,
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "type is required, when no push_branch defined",
			triggerMapItem: TriggerMapItemModel{
				Tag:        "",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "it fails when not tag related condition provided - commit message",
			triggerMapItem: TriggerMapItemModel{
				Tag:           "1.0",
				CommitMessage: "msg",
				WorkflowID:    "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both tag and commit_message defined",
		},
		{
			name: "it fails when not tag related condition provided - PR comment",
			triggerMapItem: TriggerMapItemModel{
				Tag:                "1.0",
				PullRequestComment: "msg",
				WorkflowID:         "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both tag and pull_request_comment defined",
		},
		{
			name: "it fails when not tag related condition provided - PR label",
			triggerMapItem: TriggerMapItemModel{
				Tag:              "1.0",
				PullRequestLabel: "label",
				WorkflowID:       "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both tag and pull_request_label defined",
		},
		{
			name: "it fails when not tag related condition provided - changed files",
			triggerMapItem: TriggerMapItemModel{
				Tag:          "1.0",
				ChangedFiles: "file",
				WorkflowID:   "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both tag and changed_files defined",
		},
		{
			name: "it fails for invalid code-push trigger item - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				Tag: "*",
			},
			wantErr: "trigger item #1: no pipeline nor workflow is defined as trigger target",
		},
		{
			name: "tag can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Tag: map[interface{}]interface{}{
					"regex": "feature-.*",
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "regex with negative lookahead is supported",
			triggerMapItem: TriggerMapItemModel{
				Tag: map[interface{}]interface{}{
					"regex": "^(?!.SprintPreparationJob).$",
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
		},
		{
			name: "regex is validated",
			triggerMapItem: TriggerMapItemModel{
				Tag: map[interface{}]interface{}{
					"regex": "(invalid expression",
				},
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: invalid regex value in tag field: end pattern with unmatched parenthesis",
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

func TestTriggerMapItemModel_Validate_PullRequestItem(t *testing.T) {
	tests := []struct {
		name           string
		triggerMapItem TriggerMapItemModel
		workflows      []string
		pipelines      []string
		wantWarns      []string
		wantErr        string
	}{
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
			name: "type is required, when no condition defined",
			triggerMapItem: TriggerMapItemModel{
				Type:       PullRequestType,
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "type is required, when no pull_request_source_branch defined (pull_request_label)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestLabel: "CI",
				PipelineID:       "primary",
			},
			pipelines: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no pull_request_source_branch defined (pull_request_comment)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestComment: "comment",
				PipelineID:         "primary",
			},
			pipelines: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no pull_request_source_branch defined (commit_message)",
			triggerMapItem: TriggerMapItemModel{
				CommitMessage: "commit",
				PipelineID:    "primary",
			},
			pipelines: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no pull_request_source_branch defined (changed_files)",
			triggerMapItem: TriggerMapItemModel{
				ChangedFiles: "my file",
				PipelineID:   "primary",
			},
			pipelines: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "type is required, when no pull_request_source_branch defined (draft_pull_request_enabled)",
			triggerMapItem: TriggerMapItemModel{
				DraftPullRequestEnabled: pointers.NewBoolPtr(false),
				PipelineID:              "primary",
			},
			pipelines: []string{"primary"},
			wantErr:   "trigger item #1: no type or relevant trigger condition defined",
		},
		{
			name: "it fails for invalid pull-request trigger item (target branch set) - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "*",
			},
			wantErr: "trigger item #1: no pipeline nor workflow is defined as trigger target",
		},
		{
			name: "it fails for invalid pull-request trigger item (target and source branch set) - missing pipeline & workflow",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature*",
				PullRequestTargetBranch: "master",
			},
			wantErr: "trigger item #1: no pipeline nor workflow is defined as trigger target",
		},
		{
			name: "pull_request_source_branch can be a regex",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: map[interface{}]interface{}{
					"regex": "feature-.*",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "pull_request_target_branch can be a regex",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: map[string]interface{}{
					"regex": "feature-.*",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "pull_request_label can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: PullRequestType,
				PullRequestLabel: map[string]string{
					"regex": "CI",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "commit_message can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: PullRequestType,
				CommitMessage: map[string]string{
					"regex": "commit msg",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "changed_files can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: PullRequestType,
				ChangedFiles: map[string]string{
					"regex": "files",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "pull_request_comment can be a regex",
			triggerMapItem: TriggerMapItemModel{
				Type: PullRequestType,
				PullRequestComment: map[string]string{
					"regex": "CI",
				},
				PipelineID: "primary",
			},
			pipelines: []string{"primary"},
		},
		{
			name: "it fails for mixed type trigger item (pull_request_source_branch + tag)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "feature/*",
				Tag:                     "master",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both pull_request_source_branch and tag defined",
		},
		{
			name: "it fails for mixed type trigger item (pull_request_target_branch + tag)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestTargetBranch: "feature/*",
				Tag:                     "master",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both pull_request_target_branch and tag defined",
		},
		{
			name: "it fails for mixed type trigger item (pull_request_source_branch + pull_request_target_branch + tag)",
			triggerMapItem: TriggerMapItemModel{
				PullRequestSourceBranch: "main",
				PullRequestTargetBranch: "feature/*",
				Tag:                     "master",
				WorkflowID:              "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both pull_request_source_branch and pull_request_target_branch and tag defined",
		},
		{
			name: "it fails for mixed type trigger item (type + tag)",
			triggerMapItem: TriggerMapItemModel{
				Type:       PullRequestType,
				Tag:        "master",
				WorkflowID: "primary",
			},
			workflows: []string{"primary"},
			wantErr:   "trigger item #1: both type: pull_request and tag defined",
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
