package models

import (
	"fmt"
)

type Triggers struct {
	PushTriggers        []PushGitEventTriggerItem        `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequestTriggers []PullRequestGitEventTriggerItem `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	TagTriggers         []TagGitEventTriggerItem         `json:"tag,omitempty" yaml:"tag,omitempty"`
}

type PushGitEventTriggerItem struct {
	Enabled       *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Branch        any   `json:"branch,omitempty" yaml:"branch,omitempty"`
	CommitMessage any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type PullRequestGitEventTriggerItem struct {
	Enabled       *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	DraftEnabled  *bool `json:"draft_enabled,omitempty" yaml:"draft_enabled,omitempty"`
	SourceBranch  any   `json:"source_branch,omitempty" yaml:"source_branch,omitempty"`
	TargetBranch  any   `json:"target_branch,omitempty" yaml:"target_branch,omitempty"`
	Label         any   `json:"label,omitempty" yaml:"label,omitempty"`
	Comment       any   `json:"comment,omitempty" yaml:"comment,omitempty"`
	CommitMessage any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type TagGitEventTriggerItem struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Name    any   `json:"name,omitempty" yaml:"name,omitempty"`
}

// TODO: check error messages
func (triggers *Triggers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var triggersConfig map[string]any
	if err := unmarshal(&triggersConfig); err != nil {
		return fmt.Errorf("'triggers' should be an object with 'push', 'pull_request' and 'tag' keys")
	}

	if pushTriggersRaw, ok := triggersConfig["push"]; ok {
		pushTriggers, err := parsePushTriggers(pushTriggersRaw)
		if err != nil {
			return err
		}

		triggers.PushTriggers = pushTriggers
	}

	if pullRequestTriggersRaw, ok := triggersConfig["pull_request"]; ok {
		pullRequestTriggers, err := parsePullRequestTriggers(pullRequestTriggersRaw)
		if err != nil {
			return err
		}

		triggers.PullRequestTriggers = pullRequestTriggers
	}

	if tagTriggersRaw, ok := triggersConfig["tag"]; ok {
		tagTriggers, err := parseTagTriggers(tagTriggersRaw)
		if err != nil {
			return err
		}

		triggers.TagTriggers = tagTriggers
	}

	return nil
}

func parsePushTriggers(pushTriggersRaw any) ([]PushGitEventTriggerItem, error) {
	pushTriggersList, ok := pushTriggersRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.push' should be a list of push trigger items")
	}

	var pushTriggers []PushGitEventTriggerItem
	for idx, pushTriggerRaw := range pushTriggersList {
		pushTriggerItem, err := parsePushTriggerItem(pushTriggerRaw, idx)
		if err != nil {
			return nil, err
		}

		pushTriggers = append(pushTriggers, *pushTriggerItem)
	}

	return pushTriggers, nil
}

func parsePullRequestTriggers(pullRequestTriggersRaw any) ([]PullRequestGitEventTriggerItem, error) {
	pullRequestTriggersList, ok := pullRequestTriggersRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.pull_request' should be a list of pull request trigger items")
	}

	var pullRequestTriggers []PullRequestGitEventTriggerItem
	for idx, pullRequestTriggerRaw := range pullRequestTriggersList {
		pullRequestTriggerItem, err := parsePullRequestTriggerItem(pullRequestTriggerRaw, idx)
		if err != nil {
			return nil, err
		}

		pullRequestTriggers = append(pullRequestTriggers, *pullRequestTriggerItem)
	}

	return pullRequestTriggers, nil
}

func parseTagTriggers(tagTriggersRaw any) ([]TagGitEventTriggerItem, error) {
	tagTriggersList, ok := tagTriggersRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.tag' should be a list of tag trigger items")
	}

	var tagTriggers []TagGitEventTriggerItem
	for idx, tagTriggerRaw := range tagTriggersList {
		tagTriggerItem, err := parseTagTriggerItem(tagTriggerRaw, idx)
		if err != nil {
			return nil, err
		}

		tagTriggers = append(tagTriggers, *tagTriggerItem)
	}

	return tagTriggers, nil
}

func parsePushTriggerItem(pushTriggerRaw any, idx int) (*PushGitEventTriggerItem, error) {
	pushTrigger, ok := pushTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.push[%d]' should be an object with 'enabled', 'branch', 'commit_message' and 'changed_files' keys", idx)
	}

	enabled, err := boolPtrValue(pushTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	branch, err := globOrRegexValue(pushTrigger, "branch")
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
		Branch:        branch,
		CommitMessage: commitMessage,
		ChangedFiles:  changedFiles,
	}, nil
}

func parsePullRequestTriggerItem(pullRequestTriggerRaw any, idx int) (*PullRequestGitEventTriggerItem, error) {
	pullRequestTrigger, ok := pullRequestTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.pull_request[%d]' should be an object with 'enabled', 'source_branch', 'target_branch', 'draft_enabled', 'label', 'comment', 'commit_message' and 'changed_files' keys", idx)
	}

	enabled, err := boolPtrValue(pullRequestTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	draftEnabled, err := boolPtrValue(pullRequestTrigger, "draft_enabled")
	if err != nil {
		return nil, err
	}

	sourceBranch, err := globOrRegexValue(pullRequestTrigger, "source_branch")
	if err != nil {
		return nil, err
	}

	targetBranch, err := globOrRegexValue(pullRequestTrigger, "target_branch")
	if err != nil {
		return nil, err
	}

	label, err := globOrRegexValue(pullRequestTrigger, "label")
	if err != nil {
		return nil, err
	}

	comment, err := globOrRegexValue(pullRequestTrigger, "comment")
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
		Enabled:       enabled,
		SourceBranch:  sourceBranch,
		TargetBranch:  targetBranch,
		DraftEnabled:  draftEnabled,
		Label:         label,
		Comment:       comment,
		CommitMessage: commitMessage,
		ChangedFiles:  changedFiles,
	}, nil
}

func parseTagTriggerItem(tagTriggerRaw any, idx int) (*TagGitEventTriggerItem, error) {
	tagTrigger, ok := tagTriggerRaw.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.tag[%d]' should be an object with 'enabled' and 'name' keys", idx)
	}

	enabled, err := boolPtrValue(tagTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	name, err := globOrRegexValue(tagTrigger, "name")
	if err != nil {
		return nil, err
	}

	return &TagGitEventTriggerItem{
		Enabled: enabled,
		Name:    name,
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
