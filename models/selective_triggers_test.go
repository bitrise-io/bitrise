package models

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalTriggersYAML(t *testing.T) {
	tests := []struct {
		name         string
		yamlContent  string
		wantTriggers Triggers
	}{
		{
			name: "Parses push event trigger item with glob filters",
			yamlContent: `
git_events:
  push:
  - push_branch: branch
    commit_message: message
    changed_files: file
    enabled: false`,
			wantTriggers: triggersWithPushGitEventTriggerItems(
				PushGitEventTriggerItem{
					PushBranch:    GlobOrRegexFilterValue{Glob: "branch"},
					CommitMessage: GlobOrRegexFilterValue{Glob: "message"},
					ChangedFiles:  GlobOrRegexFilterValue{Glob: "file"},
					Enabled:       pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Parses push event trigger item with regex filters",
			yamlContent: `
git_events:
  push:
  - push_branch:
      regex: branch
    commit_message: 
      regex: message
    changed_files: 
      regex: file
    enabled: false`,
			wantTriggers: triggersWithPushGitEventTriggerItems(
				PushGitEventTriggerItem{
					PushBranch:    GlobOrRegexFilterValue{Regex: "branch"},
					CommitMessage: GlobOrRegexFilterValue{Regex: "message"},
					ChangedFiles:  GlobOrRegexFilterValue{Regex: "file"},
					Enabled:       pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Parses pull request event trigger item with glob filters",
			yamlContent: `
git_events:
  pull_request:
  - pull_request_source_branch: source_branch
    pull_request_target_branch: target_branch 
    draft_pull_request_enabled: false
    pull_request_label: label
    pull_request_comment: comment
    commit_message: message
    changed_files: file
    enabled: false`,
			wantTriggers: triggersWithPullRequestEventTriggerItems(
				PullRequestGitEventTriggerItem{
					PullRequestSourceBranch: GlobOrRegexFilterValue{Glob: "source_branch"},
					PullRequestTargetBranch: GlobOrRegexFilterValue{Glob: "target_branch"},
					DraftPullRequestEnabled: pointers.NewBoolPtr(false),
					PullRequestLabel:        GlobOrRegexFilterValue{Glob: "label"},
					PullRequestComment:      GlobOrRegexFilterValue{Glob: "comment"},
					CommitMessage:           GlobOrRegexFilterValue{Glob: "message"},
					ChangedFiles:            GlobOrRegexFilterValue{Glob: "file"},
					Enabled:                 pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Parses pull request event trigger item with regex filters",
			yamlContent: `
git_events:
  pull_request:
  - pull_request_source_branch:
      regex: source_branch
    pull_request_target_branch:
      regex: target_branch 
    draft_pull_request_enabled: false
    pull_request_label:
      regex: label
    pull_request_comment:
      regex: comment
    commit_message:
      regex: message
    changed_files:
      regex: file
    enabled: false`,
			wantTriggers: triggersWithPullRequestEventTriggerItems(
				PullRequestGitEventTriggerItem{
					PullRequestSourceBranch: GlobOrRegexFilterValue{Regex: "source_branch"},
					PullRequestTargetBranch: GlobOrRegexFilterValue{Regex: "target_branch"},
					DraftPullRequestEnabled: pointers.NewBoolPtr(false),
					PullRequestLabel:        GlobOrRegexFilterValue{Regex: "label"},
					PullRequestComment:      GlobOrRegexFilterValue{Regex: "comment"},
					CommitMessage:           GlobOrRegexFilterValue{Regex: "message"},
					ChangedFiles:            GlobOrRegexFilterValue{Regex: "file"},
					Enabled:                 pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Parses tag event trigger item with glob filters",
			yamlContent: `
git_events:
  tag:
  - tag: tag
    enabled: false`,
			wantTriggers: triggersWithTagEventTriggerItems(
				TagGitEventTriggerItem{
					Tag:     GlobOrRegexFilterValue{Glob: "tag"},
					Enabled: pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Parses tag event trigger item with regex filters",
			yamlContent: `
git_events:
  tag:
  - tag: 
      regex: tag
    enabled: false`,
			wantTriggers: triggersWithTagEventTriggerItems(
				TagGitEventTriggerItem{
					Tag:     GlobOrRegexFilterValue{Regex: "tag"},
					Enabled: pointers.NewBoolPtr(false),
				}),
		},
		{
			name: "Throws error when filter value is not a string or an object with a 'regex' key and string value",
			yamlContent: `
git_events:
  tag:
  - tag: 
      glob: tag
    enabled: false`,
			wantTriggers: triggersWithTagEventTriggerItems(
				TagGitEventTriggerItem{
					Tag:     GlobOrRegexFilterValue{Regex: "tag"},
					Enabled: pointers.NewBoolPtr(false),
				}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.UnmarshalStrict([]byte(tt.yamlContent), &triggers)
			require.NoError(t, err)
			require.EqualValues(t, tt.wantTriggers, triggers)
		})
	}

	yamlContent := `
git_events:
  push:
  - push_branch: main
    enabled: false`

	var triggers Triggers
	err := yaml.Unmarshal([]byte(yamlContent), &triggers)
	require.NoError(t, err)
	require.Equal(t, 1, len(triggers.GitEventTriggers.PushTriggers))
	require.Equal(t, "main", triggers.GitEventTriggers.PushTriggers[0].PushBranch.Glob)
}

func triggersWithPushGitEventTriggerItems(items ...PushGitEventTriggerItem) Triggers {
	return Triggers{
		GitEventTriggers: GitEventTriggers{
			PushTriggers: items,
		},
	}
}

func triggersWithPullRequestEventTriggerItems(items ...PullRequestGitEventTriggerItem) Triggers {
	return Triggers{
		GitEventTriggers: GitEventTriggers{
			PullRequestTriggers: items,
		},
	}
}

func triggersWithTagEventTriggerItems(items ...TagGitEventTriggerItem) Triggers {
	return Triggers{
		GitEventTriggers: GitEventTriggers{
			TagTriggers: items,
		},
	}
}
