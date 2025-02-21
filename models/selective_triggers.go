package models

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

type Triggers struct {
	Enabled             *bool                            `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PushTriggers        []PushGitEventTriggerItem        `json:"push,omitempty" yaml:"push,omitempty"`
	PullRequestTriggers []PullRequestGitEventTriggerItem `json:"pull_request,omitempty" yaml:"pull_request,omitempty"`
	TagTriggers         []TagGitEventTriggerItem         `json:"tag,omitempty" yaml:"tag,omitempty"`
}

type PushGitEventTriggerItem struct {
	Enabled       *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Priority      *int  `json:"priority,omitempty" yaml:"priority,omitempty"`
	Branch        any   `json:"branch,omitempty" yaml:"branch,omitempty"`
	CommitMessage any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type PullRequestGitEventTriggerItem struct {
	Enabled       *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Priority      *int  `json:"priority,omitempty" yaml:"priority,omitempty"`
	DraftEnabled  *bool `json:"draft_enabled,omitempty" yaml:"draft_enabled,omitempty"`
	SourceBranch  any   `json:"source_branch,omitempty" yaml:"source_branch,omitempty"`
	TargetBranch  any   `json:"target_branch,omitempty" yaml:"target_branch,omitempty"`
	Label         any   `json:"label,omitempty" yaml:"label,omitempty"`
	Comment       any   `json:"comment,omitempty" yaml:"comment,omitempty"`
	CommitMessage any   `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  any   `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`
}

type TagGitEventTriggerItem struct {
	Enabled  *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Priority *int  `json:"priority,omitempty" yaml:"priority,omitempty"`
	Name     any   `json:"name,omitempty" yaml:"name,omitempty"`
}

func (pushItem PushGitEventTriggerItem) toString() string {
	return fmt.Sprintf("PushGitEventTriggerItem{Branch: %v, CommitMessage: %v, ChangedFiles: %v}", pushItem.Branch, pushItem.CommitMessage, pushItem.ChangedFiles)
}

func (pullRequestItem PullRequestGitEventTriggerItem) toString() string {
	draftEnabled := defaultDraftPullRequestEnabled
	if pullRequestItem.DraftEnabled != nil {
		draftEnabled = *pullRequestItem.DraftEnabled
	}
	return fmt.Sprintf("PullRequestGitEventTriggerItem{DraftEnabled: %v, SourceBranch: %v, TargetBranch: %v, Label: %v, Comment: %v, CommitMessage: %v, ChangedFiles: %v}", draftEnabled, pullRequestItem.SourceBranch, pullRequestItem.TargetBranch, pullRequestItem.Label, pullRequestItem.Comment, pullRequestItem.CommitMessage, pullRequestItem.ChangedFiles)
}

func (tagItem TagGitEventTriggerItem) toString() string {
	return fmt.Sprintf("TagGitEventTriggerItem{Name: %v}", tagItem.Name)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Triggers, allowing
// additional validation of the triggers YAML configuration.
func (triggers *Triggers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var triggersConfig map[string]any
	if err := unmarshal(&triggersConfig); err != nil {
		return fmt.Errorf("'triggers': should be a map with 'enabled', 'push', 'pull_request' and 'tag' keys")
	}

	if err := ensureKeys(triggersConfig, "enabled", "push", "pull_request", "tag"); err != nil {
		return fmt.Errorf("'triggers': %w", err)
	}

	enabled, err := boolPtrValue(triggersConfig, "enabled")
	if err != nil {
		return fmt.Errorf("'triggers': %w", err)
	}
	triggers.Enabled = enabled

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
		return nil, fmt.Errorf("'triggers.push': should be a list of push trigger items")
	}

	seenPushItems := map[string]int{}
	var pushTriggers []PushGitEventTriggerItem
	for idx, pushTriggerRaw := range pushTriggersList {
		pushTriggerItem, err := parsePushTriggerItem(pushTriggerRaw)
		if err != nil {
			return nil, fmt.Errorf("'triggers.push[%d]': %w", idx, err)
		}

		pushTriggerItemStr := pushTriggerItem.toString()
		seenIdx, ok := seenPushItems[pushTriggerItemStr]
		if ok {
			return nil, fmt.Errorf("'triggers.push[%d]': duplicates push trigger item #%d", idx, seenIdx)
		}
		seenPushItems[pushTriggerItemStr] = idx

		pushTriggers = append(pushTriggers, *pushTriggerItem)
	}

	return pushTriggers, nil
}

func parsePullRequestTriggers(pullRequestTriggersRaw any) ([]PullRequestGitEventTriggerItem, error) {
	pullRequestTriggersList, ok := pullRequestTriggersRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.pull_request': should be a list of pull request trigger items")
	}

	seenPullRequestItems := map[string]int{}
	var pullRequestTriggers []PullRequestGitEventTriggerItem
	for idx, pullRequestTriggerRaw := range pullRequestTriggersList {
		pullRequestTriggerItem, err := parsePullRequestTriggerItem(pullRequestTriggerRaw)
		if err != nil {
			return nil, fmt.Errorf("'triggers.pull_request[%d]': %w", idx, err)
		}

		pullRequestTriggerItemStr := pullRequestTriggerItem.toString()
		seenIdx, ok := seenPullRequestItems[pullRequestTriggerItemStr]
		if ok {
			return nil, fmt.Errorf("'triggers.pull_request[%d]': duplicates pull request trigger item #%d", idx, seenIdx)
		}
		seenPullRequestItems[pullRequestTriggerItemStr] = idx

		pullRequestTriggers = append(pullRequestTriggers, *pullRequestTriggerItem)
	}

	return pullRequestTriggers, nil
}

func parseTagTriggers(tagTriggersRaw any) ([]TagGitEventTriggerItem, error) {
	tagTriggersList, ok := tagTriggersRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("'triggers.tag': should be a list of tag trigger items")
	}

	seenTagItems := map[string]int{}
	var tagTriggers []TagGitEventTriggerItem
	for idx, tagTriggerRaw := range tagTriggersList {
		tagTriggerItem, err := parseTagTriggerItem(tagTriggerRaw)
		if err != nil {
			return nil, fmt.Errorf("'triggers.tag[%d]': %w", idx, err)
		}

		tagTriggerItemStr := tagTriggerItem.toString()
		seenIdx, ok := seenTagItems[tagTriggerItemStr]
		if ok {
			return nil, fmt.Errorf("'triggers.tag[%d]': duplicates tag trigger item #%d", idx, seenIdx)
		}
		seenTagItems[tagTriggerItemStr] = idx

		tagTriggers = append(tagTriggers, *tagTriggerItem)
	}

	return tagTriggers, nil
}

func parsePushTriggerItem(pushTriggerRaw any) (*PushGitEventTriggerItem, error) {
	pushTrigger, ok := pushTriggerRaw.(map[any]any)
	if !ok {
		return nil, errors.New("should be a map with 'enabled', 'branch', 'commit_message' and 'changed_files' keys")
	}

	stringKeyedPushTrigger, err := stringKeyedMap(pushTrigger)
	if err != nil {
		return nil, err
	}

	if err := ensureKeys(stringKeyedPushTrigger, "enabled", "priority", "branch", "commit_message", "changed_files"); err != nil {
		return nil, err
	}

	enabled, err := boolPtrValue(stringKeyedPushTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	priority, err := priorityValue(stringKeyedPushTrigger)
	if err != nil {
		return nil, err
	}

	branch, err := globOrRegexValue(stringKeyedPushTrigger, "branch")
	if err != nil {
		return nil, err
	}

	commitMessage, err := globOrRegexValue(stringKeyedPushTrigger, "commit_message")
	if err != nil {
		return nil, err
	}

	changedFiles, err := globOrRegexValue(stringKeyedPushTrigger, "changed_files")
	if err != nil {
		return nil, err
	}

	return &PushGitEventTriggerItem{
		Enabled:       enabled,
		Priority:      priority,
		Branch:        branch,
		CommitMessage: commitMessage,
		ChangedFiles:  changedFiles,
	}, nil
}

func parsePullRequestTriggerItem(pullRequestTriggerRaw any) (*PullRequestGitEventTriggerItem, error) {
	pullRequestTrigger, ok := pullRequestTriggerRaw.(map[any]any)
	if !ok {
		return nil, errors.New("should be a map with 'enabled', 'source_branch', 'target_branch', 'draft_enabled', 'label', 'comment', 'commit_message' and 'changed_files' keys")
	}

	stringKeyedPullRequestTrigger, err := stringKeyedMap(pullRequestTrigger)
	if err != nil {
		return nil, err
	}

	if err := ensureKeys(stringKeyedPullRequestTrigger, "enabled", "priority", "source_branch", "target_branch", "draft_enabled", "label", "comment", "commit_message", "changed_files"); err != nil {
		return nil, err
	}

	enabled, err := boolPtrValue(stringKeyedPullRequestTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	priority, err := priorityValue(stringKeyedPullRequestTrigger)
	if err != nil {
		return nil, err
	}

	draftEnabled, err := boolPtrValue(stringKeyedPullRequestTrigger, "draft_enabled")
	if err != nil {
		return nil, err
	}

	sourceBranch, err := globOrRegexValue(stringKeyedPullRequestTrigger, "source_branch")
	if err != nil {
		return nil, err
	}

	targetBranch, err := globOrRegexValue(stringKeyedPullRequestTrigger, "target_branch")
	if err != nil {
		return nil, err
	}

	label, err := globOrRegexValue(stringKeyedPullRequestTrigger, "label")
	if err != nil {
		return nil, err
	}

	comment, err := globOrRegexValue(stringKeyedPullRequestTrigger, "comment")
	if err != nil {
		return nil, err
	}

	commitMessage, err := globOrRegexValue(stringKeyedPullRequestTrigger, "commit_message")
	if err != nil {
		return nil, err
	}

	changedFiles, err := globOrRegexValue(stringKeyedPullRequestTrigger, "changed_files")
	if err != nil {
		return nil, err
	}

	return &PullRequestGitEventTriggerItem{
		Enabled:       enabled,
		Priority:      priority,
		SourceBranch:  sourceBranch,
		TargetBranch:  targetBranch,
		DraftEnabled:  draftEnabled,
		Label:         label,
		Comment:       comment,
		CommitMessage: commitMessage,
		ChangedFiles:  changedFiles,
	}, nil
}

func parseTagTriggerItem(tagTriggerRaw any) (*TagGitEventTriggerItem, error) {
	tagTrigger, ok := tagTriggerRaw.(map[any]any)
	if !ok {
		return nil, errors.New("should be a map with 'enabled' and 'name' keys")
	}

	stringKeyedTagTrigger, err := stringKeyedMap(tagTrigger)
	if err != nil {
		return nil, err
	}

	if err := ensureKeys(stringKeyedTagTrigger, "enabled", "priority", "name"); err != nil {
		return nil, err
	}

	enabled, err := boolPtrValue(stringKeyedTagTrigger, "enabled")
	if err != nil {
		return nil, err
	}

	priority, err := priorityValue(stringKeyedTagTrigger)
	if err != nil {
		return nil, err
	}

	name, err := globOrRegexValue(stringKeyedTagTrigger, "name")
	if err != nil {
		return nil, err
	}

	return &TagGitEventTriggerItem{
		Enabled:  enabled,
		Priority: priority,
		Name:     name,
	}, nil
}

func globOrRegexValue(item map[string]any, key string) (any, error) {
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
			return nil, fmt.Errorf("'%s' value should be a string or a map with a 'regex' key and string value", key)
		}
		return map[string]string{"regex": regex}, nil
	default:
		return nil, fmt.Errorf("'%s' value should be a string or a map with a 'regex' key and string value", key)
	}
}

func boolPtrValue(item map[string]any, key string) (*bool, error) {
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
func priorityValue(item map[string]any) (*int, error) {
	valuePtr, err := intPtrValue(item, "priority")
	if err != nil {
		return nil, err
	}
	if err := validatePriority(valuePtr); err != nil {
		return nil, err
	}
	return valuePtr, nil
}

func intPtrValue(item map[string]any, key string) (*int, error) {
	value, ok := item[key]
	if !ok {
		return nil, nil
	}

	intValue, ok := value.(int)
	if !ok {
		return nil, fmt.Errorf("'%s' value should be an integer", key)
	}

	return &intValue, nil
}

func ensureKeys(item map[string]any, allowedKeys ...string) error {
	keys := maps.Keys(item)
	for _, allowedKey := range allowedKeys {
		idx := slices.Index(keys, allowedKey)
		if idx >= 0 {
			keys = slices.Delete(keys, idx, idx+1)
		}
	}
	if len(keys) > 0 {
		return fmt.Errorf("unknown key(s): %s", strings.Join(keys, ", "))
	}

	return nil
}

func stringKeyedMap(item map[any]any) (map[string]any, error) {
	stringKeyedItem := make(map[string]any, len(item))
	for key, value := range item {
		keyStr, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("should be a string keyed map")
		}
		stringKeyedItem[keyStr] = value
	}
	return stringKeyedItem, nil
}
