package models

import (
	"fmt"
)

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

// TODO: check error messages
func (triggers *Triggers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var triggersConfig map[string]any
	if err := unmarshal(&triggersConfig); err != nil {
		return fmt.Errorf("triggers should be an object with git_events key")
	}

	gitEventTriggersRaw, ok := triggersConfig["git_events"]
	if ok {
		gitEventTriggers, err := parseGitEventTriggers(gitEventTriggersRaw)
		if err != nil {
			return err
		}

		triggers.GitEventTriggers = *gitEventTriggers
	}

	return nil
}

func parseGitEventTriggers(gitEventTriggersRaw any) (*GitEventTriggers, error) {
	gitEventTriggersConfig, ok := gitEventTriggersRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("git_events should be an object with push, pull_request and tag keys")
	}

	var pushTriggers []PushGitEventTriggerItem
	pushTriggersRaw := gitEventTriggersConfig["push"]
	if pushTriggersRaw != nil {
		pushTriggersList, ok := pushTriggersRaw.([]any)
		if !ok {
			return nil, fmt.Errorf("push trigger should be a list of objects")
		}

		for _, pushTriggerRaw := range pushTriggersList {
			pushTriggerItem, err := parsePushTriggerItem(pushTriggerRaw)
			if err != nil {
				return nil, err
			}

			pushTriggers = append(pushTriggers, *pushTriggerItem)
		}
	}

	var pullRequestTriggers []PullRequestGitEventTriggerItem
	pullRequestTriggersRaw := gitEventTriggersConfig["pull_request"]
	if pullRequestTriggersRaw != nil {
		pullRequestTriggersList := pullRequestTriggersRaw.([]any)
		for _, pullRequestTriggerRaw := range pullRequestTriggersList {
			pullRequestTriggerItem, err := parsePullRequestTriggerItem(pullRequestTriggerRaw)
			if err != nil {
				return nil, err
			}

			pullRequestTriggers = append(pullRequestTriggers, *pullRequestTriggerItem)
		}
	}

	var tagTriggers []TagGitEventTriggerItem
	tagTriggersRaw := gitEventTriggersConfig["tag"]
	if tagTriggersRaw != nil {
		tagTriggersList := tagTriggersRaw.([]any)
		for _, tagTriggerRaw := range tagTriggersList {
			tagTriggerItem, err := parseTagTriggerItem(tagTriggerRaw)
			if err != nil {
				return nil, err
			}

			tagTriggers = append(tagTriggers, *tagTriggerItem)
		}
	}

	return &GitEventTriggers{
		PushTriggers:        pushTriggers,
		PullRequestTriggers: pullRequestTriggers,
		TagTriggers:         tagTriggers,
	}, nil
}

func parsePushTriggerItem(pushTriggerRaw any) (*PushGitEventTriggerItem, error) {
	pushTrigger, ok := pushTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("push trigger should be an object with enabled, push_branch, commit_message and changed_files keys")
	}

	enabled, err := boolPtrValue(pushTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	pushBranch, err := globOrRegexValue(pushTrigger, "push_branch")
	if err != nil {
		return nil, err
	}

	commitMessage, err := globOrRegexValue(pushTrigger, "commit_message")
	if err != nil {
		return nil, err
	}

	changedFiles, err := globOrRegexValue(pushTrigger, "changed_files")
	if err != nil {
		return nil, err
	}

	return &PushGitEventTriggerItem{
		Enabled:       enabled,
		PushBranch:    pushBranch,
		CommitMessage: commitMessage,
		ChangedFiles:  changedFiles,
	}, nil
}

func parsePullRequestTriggerItem(pullRequestTriggerRaw any) (*PullRequestGitEventTriggerItem, error) {
	pullRequestTrigger, ok := pullRequestTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("pull request trigger should be an object with enabled, pull_request_source_branch, pull_request_target_branch, draft_pull_request_enabled, pull_request_label, pull_request_comment, commit_message and changed_files keys")
	}

	enabled, err := boolPtrValue(pullRequestTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	draftPullRequestEnabled, err := boolPtrValue(pullRequestTrigger, "draft_pull_request_enabled")
	if err != nil {
		return nil, err
	}

	pullRequestSourceBranch, err := globOrRegexValue(pullRequestTrigger, "pull_request_source_branch")
	if err != nil {
		return nil, err
	}

	pullRequestTargetBranch, err := globOrRegexValue(pullRequestTrigger, "pull_request_target_branch")
	if err != nil {
		return nil, err
	}

	pullRequestLabel, err := globOrRegexValue(pullRequestTrigger, "pull_request_label")
	if err != nil {
		return nil, err
	}

	pullRequestComment, err := globOrRegexValue(pullRequestTrigger, "pull_request_comment")
	if err != nil {
		return nil, err
	}

	commitMessage, err := globOrRegexValue(pullRequestTrigger, "commit_message")
	if err != nil {
		return nil, err
	}

	changedFiles, err := globOrRegexValue(pullRequestTrigger, "changed_files")
	if err != nil {
		return nil, err
	}

	return &PullRequestGitEventTriggerItem{
		Enabled:                 enabled,
		PullRequestSourceBranch: pullRequestSourceBranch,
		PullRequestTargetBranch: pullRequestTargetBranch,
		DraftPullRequestEnabled: draftPullRequestEnabled,
		PullRequestLabel:        pullRequestLabel,
		PullRequestComment:      pullRequestComment,
		CommitMessage:           commitMessage,
		ChangedFiles:            changedFiles,
	}, nil
}

func parseTagTriggerItem(tagTriggerRaw any) (*TagGitEventTriggerItem, error) {
	tagTrigger, ok := tagTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("tag trigger should be an object with enabled and tag keys")
	}

	enabled, err := boolPtrValue(tagTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	tag, err := globOrRegexValue(tagTrigger, "tag")
	if err != nil {
		return nil, err
	}

	return &TagGitEventTriggerItem{
		Enabled: enabled,
		Tag:     tag,
	}, nil
}

func globOrRegexValue(item map[any]any, key string) (any, error) {
	value, ok := item[key]
	if !ok {
		return nil, nil
	}

	switch value := value.(type) {
	case string:
		return value, nil
	case map[any]any:
		regexRaw := value["regex"]
		regex, ok := regexRaw.(string)
		if !ok {
			return nil, fmt.Errorf("'%s' value should be a string or an object with a 'regex' key and string value", key)
		}
		return map[string]string{"regex": regex}, nil
	default:
		return nil, fmt.Errorf("'%s' value should be a string or an object with a 'regex' key and string value", key)
	}
}

func boolPtrValue(item map[any]any, key string) (*bool, error) {
	value, ok := item[key]
	if !ok {
		return nil, nil
	}

	boolValue, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("'%s' value should be a boolean", key)
	}

	return &boolValue, nil
}
