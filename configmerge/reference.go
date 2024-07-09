package configmerge

import (
	"fmt"
	"path/filepath"
)

type ConfigReference struct {
	Repository string `yaml:"repository" json:"repository"`
	Branch     string `yaml:"branch" json:"branch"`
	Commit     string `yaml:"commit" json:"commit"`
	Tag        string `yaml:"tag" json:"tag"`
	Path       string `yaml:"path" json:"path"`
}

func NewConfigReference(repository, branch, commit, tag, path string) ConfigReference {
	return ConfigReference{
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Tag:        tag,
	}
}

func (r ConfigReference) Key() string {
	if r.Branch == "" && r.Tag == "" && r.Commit == "" {
		return ""
	}

	key := r.Path
	if r.Repository != "" {
		key = "repo:" + r.Repository + "," + r.Path
	}

	if r.Commit != "" {
		key += "@commit:" + r.Commit
	} else if r.Tag != "" {
		key += "@tag:" + r.Tag
	} else if r.Branch != "" {
		key += "@branch:" + r.Branch
	}

	return key
}

func (r ConfigReference) Validate() error {
	key := r.Key()

	includePath := r.Path
	if includePath == "" {
		return fmt.Errorf("missing YML path in reference: %s", key)
	}

	if filepath.Ext(includePath) != ".yml" && filepath.Ext(includePath) != ".yaml" {
		return fmt.Errorf("invalid YML path in reference (%s): %s is not a yaml file", key, includePath)
	}

	includeCommit := r.Commit
	isCommitValid := true
	if includeCommit != "" {
		isCommitValid = false

		if len(includeCommit) > 5 && len(includeCommit) < 9 {
			isCommitValid = true
		} else if len(includeCommit) == 40 {
			isCommitValid = true
		}
	}
	if !isCommitValid {
		return fmt.Errorf("invalid commit hash in reference (%s): %s", key, includeCommit)
	}

	includeRepo := r.Repository
	includeBranch := r.Branch
	includeTag := r.Tag
	if includeRepo != "" && includeBranch == "" && includeTag == "" && includeCommit == "" {
		return fmt.Errorf("incomplete reference (%s): repository specified without branch, tag or commit", key)

	}

	return nil
}
