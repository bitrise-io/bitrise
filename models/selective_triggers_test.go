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
		wantErr      string
	}{
		{
			name: "Parses push event trigger item with glob filters",
			yamlContent: `
push:
- branch: branch
  commit_message: message
  changed_files: file
  enabled: false`,
			wantTriggers: Triggers{PushTriggers: []PushGitEventTriggerItem{{
				Branch:        "branch",
				CommitMessage: "message",
				ChangedFiles:  "file",
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses push event trigger item with regex filters",
			yamlContent: `
push:
- branch:
    regex: branch
  commit_message: 
    regex: message
  changed_files: 
    regex: file
  enabled: false`,
			wantTriggers: Triggers{PushTriggers: []PushGitEventTriggerItem{{
				Branch:        map[string]string{"regex": "branch"},
				CommitMessage: map[string]string{"regex": "message"},
				ChangedFiles:  map[string]string{"regex": "file"},
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses pull request event trigger item with glob filters",
			yamlContent: `
pull_request:
- source_branch: source_branch
  target_branch: target_branch 
  draft_enabled: false
  label: label
  comment: comment
  commit_message: message
  changed_files: file
  enabled: false`,
			wantTriggers: Triggers{PullRequestTriggers: []PullRequestGitEventTriggerItem{{
				SourceBranch:  "source_branch",
				TargetBranch:  "target_branch",
				DraftEnabled:  pointers.NewBoolPtr(false),
				Label:         "label",
				Comment:       "comment",
				CommitMessage: "message",
				ChangedFiles:  "file",
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses pull request event trigger item with regex filters",
			yamlContent: `
pull_request:
- source_branch:
    regex: source_branch
  target_branch:
    regex: target_branch 
  draft_enabled: false
  label:
    regex: label
  comment:
    regex: comment
  commit_message:
    regex: message
  changed_files:
    regex: file
  enabled: false`,
			wantTriggers: Triggers{PullRequestTriggers: []PullRequestGitEventTriggerItem{{
				SourceBranch:  map[string]string{"regex": "source_branch"},
				TargetBranch:  map[string]string{"regex": "target_branch"},
				DraftEnabled:  pointers.NewBoolPtr(false),
				Label:         map[string]string{"regex": "label"},
				Comment:       map[string]string{"regex": "comment"},
				CommitMessage: map[string]string{"regex": "message"},
				ChangedFiles:  map[string]string{"regex": "file"},
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses tag event trigger item with glob filters",
			yamlContent: `
tag:
- name: tag
  enabled: false`,
			wantTriggers: Triggers{TagTriggers: []TagGitEventTriggerItem{{
				Name:    "tag",
				Enabled: pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses tag event trigger item with regex filters",
			yamlContent: `
tag:
- name: 
    regex: tag
  enabled: false`,
			wantTriggers: Triggers{TagTriggers: []TagGitEventTriggerItem{{
				Name:    map[string]string{"regex": "tag"},
				Enabled: pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Throws error when filter value is not a string or an object with a 'regex' key and string value",
			yamlContent: `
tag:
- name: 
    glob: tag
  enabled: false`,
			wantErr: "'name' value should be a string or an object with a 'regex' key and string value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.Unmarshal([]byte(tt.yamlContent), &triggers)
			if tt.wantErr != "" {
				require.Errorf(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tt.wantTriggers, triggers)
			}
		})
	}
}
