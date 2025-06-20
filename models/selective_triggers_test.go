package models

import (
	"encoding/json"
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestYAMLUnmarshalTriggers(t *testing.T) {
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
  priority: 100
  enabled: false`,
			wantTriggers: Triggers{PushTriggers: []PushGitEventTriggerItem{{
				Branch:        "branch",
				CommitMessage: "message",
				ChangedFiles:  "file",
				Priority:      pointers.NewIntPtr(100),
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
				CommitMessage: map[string]any{"regex": "message"},
				ChangedFiles:  map[string]any{"regex": "file"},
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses push event trigger item with glob filters for last commit",
			yamlContent: `
push:
- branch: branch
  commit_message:
    pattern: message
    last_commit: true
  changed_files:
    pattern: file
    last_commit: false
  priority: 100
  enabled: false`,
			wantTriggers: Triggers{PushTriggers: []PushGitEventTriggerItem{{
				Branch:        "branch",
				CommitMessage: map[string]any{"pattern": "message", "last_commit": true},
				ChangedFiles:  map[string]any{"pattern": "file", "last_commit": false},
				Priority:      pointers.NewIntPtr(100),
				Enabled:       pointers.NewBoolPtr(false)},
			}},
		},
		{
			name: "Parses push event trigger item with regex filters for last commit",
			yamlContent: `
push:
- branch: branch
  commit_message:
    regex: message
    last_commit: true
  changed_files:
    regex: file
    last_commit: false
  priority: 100
  enabled: false`,
			wantTriggers: Triggers{PushTriggers: []PushGitEventTriggerItem{{
				Branch:        "branch",
				CommitMessage: map[string]any{"regex": "message", "last_commit": true},
				ChangedFiles:  map[string]any{"regex": "file", "last_commit": false},
				Priority:      pointers.NewIntPtr(100),
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
  priority: 100
  enabled: false`,
			wantTriggers: Triggers{PullRequestTriggers: []PullRequestGitEventTriggerItem{{
				SourceBranch:  "source_branch",
				TargetBranch:  "target_branch",
				DraftEnabled:  pointers.NewBoolPtr(false),
				Label:         "label",
				Comment:       "comment",
				CommitMessage: "message",
				ChangedFiles:  "file",
				Priority:      pointers.NewIntPtr(100),
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
  priority: 100
  enabled: false`,
			wantTriggers: Triggers{TagTriggers: []TagGitEventTriggerItem{{
				Name:     "tag",
				Priority: pointers.NewIntPtr(100),
				Enabled:  pointers.NewBoolPtr(false)},
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
			wantErr: "'triggers.tag[0]': 'name' value should be a string or a map with a 'regex' key and string value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.Unmarshal([]byte(tt.yamlContent), &triggers)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tt.wantTriggers, triggers)
			}
		})
	}
}

func TestYAMLUnmarshalTriggers_Validation_Push(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name: "Throws error when 'triggers' is not a map",
			yamlContent: `
- push_branch: "*"
  workflow: primary
- pull_request_source_branch: "*"
  workflow: primary
- tag: "*.*.*"`,
			wantErr: "'triggers': should be a map with 'enabled', 'push', 'pull_request' and 'tag' keys",
		},
		{
			name: "Throws error when 'triggers' has unknown keys",
			yamlContent: `
chron: 
- "0 0 * * *"`,
			wantErr: "'triggers': unknown key(s): chron",
		},
		{
			name: "Push should be a list of push trigger items",
			yamlContent: `
push: 
  branch: main`,
			wantErr: "'triggers.push': should be a list of push trigger items",
		},
		{
			name: "Push item should be a map",
			yamlContent: `
push: 
- main`,
			wantErr: "'triggers.push[0]': should be a map with 'enabled', 'branch', 'commit_message' and 'changed_files' keys",
		},
		{
			name: "Push item with unknown key",
			yamlContent: `
push: 
- push_branch: main`,
			wantErr: "'triggers.push[0]': unknown key(s): push_branch",
		},
		{
			name: "Push filter should be a string or a map with a 'regex' key and string value",
			yamlContent: `
push: 
- branch:
    include: main`,
			wantErr: "'triggers.push[0]': 'branch' value should be a string or a map with a 'regex' key and string value",
		},
		{
			name: "Push filter should not contain unknown keys",
			yamlContent: `
push: 
- commit_message:
    pattern: match*
    scope: 'all_commits'`,
			wantErr: "'triggers.push[0]': 'commit_message': unknown key(s): scope",
		},
		{
			name: "Push filter should not specify both 'pattern' and 'regex'",
			yamlContent: `
push: 
- commit_message:
    pattern: match*
    regex: match.*`,
			wantErr: "'triggers.push[0]': 'commit_message' should contain exactly one of 'regex' and 'pattern' keys",
		},
		{
			name: "Push filter should contain valid pattern",
			yamlContent: `
push: 
- commit_message:
    pattern: 23`,
			wantErr: "'triggers.push[0]': 'pattern' value invalid for 'commit_message', should be a string",
		},
		{
			name: "Push filter should contain valid regex",
			yamlContent: `
push: 
- commit_message:
    regex: false`,
			wantErr: "'triggers.push[0]': 'regex' value invalid for 'commit_message', should be a string",
		},
		{
			name: "Push filter should contain valid last_commit",
			yamlContent: `
push: 
- commit_message:
    pattern: something
    last_commit: "only"`,
			wantErr: "'triggers.push[0]': 'last_commit' value invalid for 'commit_message', should be a bool",
		},
		{
			name: "Duplicated push trigger items - string filters",
			yamlContent: `
push: 
- branch: main
- branch: main`,
			wantErr: "'triggers.push[1]': duplicates push trigger item #0",
		},
		{
			name: "Duplicated push trigger items - regex filters",
			yamlContent: `
push: 
- branch:
    regex: branch
- branch:
    regex: branch`,
			wantErr: "'triggers.push[1]': duplicates push trigger item #0",
		},
		{
			name: "Duplicated push trigger items - enabled",
			yamlContent: `
push: 
- branch: main
- branch: main
  enabled: false`,
			wantErr: "'triggers.push[1]': duplicates push trigger item #0",
		},
		{
			name: "Invalid priority",
			yamlContent: `
push:
- branch: main
  priority: -101`,
			wantErr: "'triggers.push[0]': priority (-101) should be between -100 and 100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.Unmarshal([]byte(tt.yamlContent), &triggers)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestYAMLUnmarshalTriggers_Validation_PullRequest(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name: "Pull request should be a list of push trigger items",
			yamlContent: `
pull_request: 
  target_branch: main`,
			wantErr: "'triggers.pull_request': should be a list of pull request trigger items",
		},
		{
			name: "Pull request item should be a map",
			yamlContent: `
pull_request: 
- main`,
			wantErr: "'triggers.pull_request[0]': should be a map with 'enabled', 'source_branch', 'target_branch', 'draft_enabled', 'label', 'comment', 'commit_message' and 'changed_files' keys",
		},
		{
			name: "Pull request item with unknown key",
			yamlContent: `
pull_request: 
- pull_request_target_branch: main`,
			wantErr: "'triggers.pull_request[0]': unknown key(s): pull_request_target_branch",
		},
		{
			name: "Pull request filter should be a string or a map with a 'regex' key and string value",
			yamlContent: `
pull_request: 
- target_branch:
    include: main`,
			wantErr: "'triggers.pull_request[0]': 'target_branch' value should be a string or a map with a 'regex' key and string value",
		},
		{
			name: "Duplicated pull request trigger items - string filters",
			yamlContent: `
pull_request:
- source_branch: source_branch
- source_branch: source_branch`,
			wantErr: "'triggers.pull_request[1]': duplicates pull request trigger item #0",
		},
		{
			name: "Duplicated pull request trigger items - regex filters",
			yamlContent: `
pull_request:
- source_branch:
    regex: source_branch
- source_branch:
    regex: source_branch`,
			wantErr: "'triggers.pull_request[1]': duplicates pull request trigger item #0",
		},
		{
			name: "Duplicated pull request trigger items - enabled",
			yamlContent: `
pull_request: 
- source_branch: source_branch
- source_branch: source_branch
  enabled: false`,
			wantErr: "'triggers.pull_request[1]': duplicates pull request trigger item #0",
		},
		{
			name: "Duplicated pull request trigger items - draft enabled",
			yamlContent: `
pull_request: 
- source_branch: source_branch
- source_branch: source_branch
  draft_enabled: true`,
			wantErr: "'triggers.pull_request[1]': duplicates pull request trigger item #0",
		},
		{
			name: "Invalid priority",
			yamlContent: `
pull_request:
- source_branch: main
  priority: 101`,
			wantErr: "'triggers.pull_request[0]': priority (101) should be between -100 and 100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.Unmarshal([]byte(tt.yamlContent), &triggers)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestYAMLUnmarshalTriggers_Validation_Tag(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name: "Tag should be a list of push trigger items",
			yamlContent: `
tag: 
  name: main`,
			wantErr: "'triggers.tag': should be a list of tag trigger items",
		},
		{
			name: "Tag item should be a map",
			yamlContent: `
tag: 
- 1.0.0`,
			wantErr: "'triggers.tag[0]': should be a map with 'enabled' and 'name' keys",
		},
		{
			name: "Tag item with unknown key",
			yamlContent: `
tag: 
- tag: main`,
			wantErr: "'triggers.tag[0]': unknown key(s): tag",
		},
		{
			name: "Tag filter should be a string or a map with a 'regex' key and string value",
			yamlContent: `
tag: 
- name:
    include: main`,
			wantErr: "'triggers.tag[0]': 'name' value should be a string or a map with a 'regex' key and string value",
		},
		{
			name: "Duplicated tag trigger items - string filters",
			yamlContent: `
tag:
- name: tag
- name: tag`,
			wantErr: "'triggers.tag[1]': duplicates tag trigger item #0",
		},
		{
			name: "Duplicated tag trigger items - regex filters",
			yamlContent: `
tag:
- name: 
    regex: tag
- name: 
    regex: tag`,
			wantErr: "'triggers.tag[1]': duplicates tag trigger item #0",
		},
		{
			name: "Duplicated tag trigger items - enabled",
			yamlContent: `
tag:
- name: tag
- name: tag
  enabled: false`,
			wantErr: "'triggers.tag[1]': duplicates tag trigger item #0",
		},
		{
			name: "Invalid priority",
			yamlContent: `
tag:
- name: main
  priority: -101`,
			wantErr: "'triggers.tag[0]': priority (-101) should be between -100 and 100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggers Triggers
			err := yaml.Unmarshal([]byte(tt.yamlContent), &triggers)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestJSONMarshalTriggers_FromYAML(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			name: "Selective triggers are JSON marshallable",
			config: `
format_version: "17"
default_step_lib_source: "https://github.com/bitrise-io/bitrise-steplib.git"

workflows:
  test:
    triggers:
      enabled: false
      push:
      - branch:
          regex: branch
      pull_request:
      - source_branch: source_branch
        enabled: false`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config BitriseDataModel
			require.NoError(t, yaml.Unmarshal([]byte(tt.config), &config))

			warns, err := config.Validate()
			require.Empty(t, warns)
			require.NoError(t, err)

			_, err = json.Marshal(config)
			require.NoError(t, err)
		})
	}
}
