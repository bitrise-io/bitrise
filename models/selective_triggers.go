package models

import "fmt"

type Triggers struct {
	GitEventTriggers GitEventTriggers `json:"git_events,omitempty" yaml:"git_events,omitempty"`
}

type GitEventTriggers struct {
	PushTriggers        []PushGitEventTriggerItem        `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequestTriggers []PullRequestGitEventTriggerItem `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	TagTriggers         []TagGitEventTriggerItem         `json:"tag,omitempty" yaml:"tag,omitempty"`
}

type PushGitEventTriggerItem struct {
	Enabled       *bool                  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PushBranch    GlobOrRegexFilterValue `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	CommitMessage GlobOrRegexFilterValue `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  GlobOrRegexFilterValue `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type PullRequestGitEventTriggerItem struct {
	Enabled                 *bool                  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PullRequestSourceBranch GlobOrRegexFilterValue `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch GlobOrRegexFilterValue `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	DraftPullRequestEnabled *bool                  `json:"draft_pull_request_enabled,omitempty" yaml:"draft_pull_request_enabled,omitempty"`
	PullRequestLabel        GlobOrRegexFilterValue `json:"pull_request_label,omitempty" yaml:"pull_request_label,omitempty"`
	PullRequestComment      GlobOrRegexFilterValue `json:"pull_request_comment,omitempty" yaml:"pull_request_comment,omitempty"`
	CommitMessage           GlobOrRegexFilterValue `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles            GlobOrRegexFilterValue `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type TagGitEventTriggerItem struct {
	Enabled *bool                  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Tag     GlobOrRegexFilterValue `json:"tag,omitempty" yaml:"tag,omitempty"`
}

type GlobOrRegexFilterValue struct {
	Glob  string
	Regex string
}

// TODO: func (filterValue *FilterValue) UnmarshalJSON(b []byte) error
func (filterValue *GlobOrRegexFilterValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var glob string
	if err := unmarshal(&glob); err == nil {
		filterValue.Glob = glob
		return nil
	}

	var regex struct {
		Regex string `yaml:"regex"`
	}
	if err := unmarshal(&regex); err == nil && regex.Regex != "" {
		filterValue.Regex = regex.Regex
		return nil
	}

	return fmt.Errorf("filter value should be a string or an object with a 'regex' key and string value")
}

//// TODO: func (filterValue *FilterValue) MarshalJSON() ([]byte, error)
//func (filterValue *PushGitEventTriggerItem) MarshalYAML() (interface{}, error) {
//	if filterValue.Glob != "" {
//		return filterValue.Glob, nil
//	}
//
//	if filterValue.Regex != "" {
//		return map[string]string{"regex": filterValue.Regex}, nil
//	}
//
//	return nil, nil
//}
