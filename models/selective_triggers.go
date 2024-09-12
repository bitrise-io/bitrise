package models

type Triggers struct {
	GitEventTriggers GitEventTriggers `json:"git_events,omitempty" yaml:"git_events,omitempty"`
}

type GitEventTriggers struct {
	PushTriggers        []PushGitEventTriggerItem        `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequestTriggers []PullRequestGitEventTriggerItem `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	TagTriggers         []TagGitEventTriggerItem         `json:"tag,omitempty" yaml:"tag,omitempty"`
}

type PushGitEventTriggerItem struct {
	Enabled       *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PushBranch    any   `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	CommitMessage any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type PullRequestGitEventTriggerItem struct {
	Enabled                 *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PullRequestSourceBranch any   `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch any   `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	DraftPullRequestEnabled *bool `json:"draft_pull_request_enabled,omitempty" yaml:"draft_pull_request_enabled,omitempty"`
	PullRequestLabel        any   `json:"pull_request_label,omitempty" yaml:"pull_request_label,omitempty"`
	PullRequestComment      any   `json:"pull_request_comment,omitempty" yaml:"pull_request_comment,omitempty"`
	CommitMessage           any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles            any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type TagGitEventTriggerItem struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Tag     any   `json:"tag,omitempty" yaml:"tag,omitempty"`
}
