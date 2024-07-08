package configmerge

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type RepoInfo struct {
	DefaultRemoteURL string
	Branch           string
	Commit           string
	Tag              string
}

type repoInfoProvider struct {
}

func NewRepoInfoProvider() RepoInfoProvider {
	return repoInfoProvider{}
}

func (p repoInfoProvider) GetRepoInfo(repoPth string) (*RepoInfo, error) {
	repo, err := git.PlainOpen(repoPth)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	// Get branch name
	var branch string
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	}

	// Get commit hash
	commitObj, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}
	commit := commitObj.Hash.String()

	// Get tag name
	var tag string
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		obj, err := repo.TagObject(ref.Hash())
		if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
			return err
		}

		if obj.Target == head.Hash() {
			tag = ref.Name().Short()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Get remote URL
	config, err := repo.Config()
	if err != nil {
		return nil, err
	}

	remote, ok := config.Remotes[git.DefaultRemoteName]
	if !ok {
		return nil, fmt.Errorf("remote %s not exists", git.DefaultRemoteName)
	}

	if len(remote.URLs) == 0 {
		return nil, fmt.Errorf("URL not set for remote %s", git.DefaultRemoteName)
	}

	remoteURL := remote.URLs[0]

	return &RepoInfo{
		DefaultRemoteURL: remoteURL,
		Branch:           branch,
		Commit:           commit,
		Tag:              tag,
	}, nil
}
