package configmerge

import (
	"io"
	"os"

	"github.com/bitrise-io/bitrise/models"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v2"
)

type Merger struct {
	repoInfo repoInfo
}

func NewMerger(repoPth string) (*Merger, error) {
	repo, err := git.PlainOpen(repoPth)
	if err != nil {
		return nil, err
	}

	info, err := readRepoInfo(repo)
	if err != nil {
		return nil, err
	}

	return &Merger{
		repoInfo: *info,
	}, nil
}

func (m Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	f, err := os.Open(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	ref := configReference{
		Repository: m.repoInfo.defaultRemoteURL,
		Commit:     m.repoInfo.commit,
		Tag:        m.repoInfo.tag,
		Branch:     m.repoInfo.branch,
		Path:       mainConfigPth,
	}

	configTree, err := m.buildConfigTree(f, ref)
	if err != nil {
		return "", nil, err
	}

	mergedConfigContent, err := configTree.Merge()
	if err != nil {
		return "", nil, err
	}

	return mergedConfigContent, configTree, nil
}

func (m Merger) buildConfigTree(configReader io.Reader, reference configReference) (*models.ConfigFileTreeModel, error) {
	b, err := io.ReadAll(configReader)
	if err != nil {
		return nil, err
	}

	var config struct {
		Include []configReference `yaml:"include" json:"include"`
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	var includedConfigs []models.ConfigFileTreeModel
	for _, include := range config.Include {
		reader, err := openConfigModule(include, m.repoInfo)
		if err != nil {
			return nil, err
		}

		tree, err := m.buildConfigTree(reader, include)
		if err != nil {
			return nil, err
		}

		includedConfigs = append(includedConfigs, *tree)
	}

	return &models.ConfigFileTreeModel{
		Path:     reference.Key(),
		Contents: string(b),
		Includes: includedConfigs,
	}, nil
}

func openConfigModule(reference configReference, info repoInfo) (io.Reader, error) {
	localReference, err := isLocalReference(reference, info)
	if err != nil {
		return nil, err
	}

	if localReference {
		return openLocalConfigModule(reference)
	} else {
		return openRemoteConfigModule(reference)
	}
}

func isLocalReference(reference configReference, info repoInfo) (bool, error) {
	if reference.Repository == "" {
		return true, nil
	}

	refGitUrl, err := parseGitRepoURL(reference.Repository)
	if err != nil {
		return false, err
	}

	repoGitURL, err := parseGitRepoURL(info.defaultRemoteURL)
	if err != nil {
		return false, err
	}

	if !equalGitRepoURLs(refGitUrl, repoGitURL) {
		return false, nil
	}

	switch {
	case reference.Commit != "":
		return reference.Commit == info.commit ||
			reference.Commit == info.commit[:7], nil
	case reference.Tag != "":
		return reference.Tag == info.tag, nil
	case reference.Branch != "":
		return reference.Branch == info.branch, nil
	}

	return true, nil
}

func openLocalConfigModule(reference configReference) (io.Reader, error) {
	return os.Open(reference.Path)
}

func openRemoteConfigModule(reference configReference) (io.Reader, error) {
	opts := git.CloneOptions{
		URL: reference.Repository,
	}
	if reference.Branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(reference.Branch)
	}

	repo, cloneErr := git.Clone(memory.NewStorage(), memfs.New(), &opts)
	if cloneErr != nil {
		if !isHttpFormatRepoURL(reference.Repository) {
			return nil, cloneErr
		}

		// Try repo url with ssh syntax
		repoURL, err := parseGitRepoURL(reference.Repository)
		if err != nil {
			return nil, err
		}
		if repoURL.User == "" {
			repoURL.User = "git"
		}

		opts := git.CloneOptions{
			URL: generateSCPStyleSSHFormatRepoURL(repoURL),
		}
		if reference.Branch != "" {
			opts.ReferenceName = plumbing.NewBranchReferenceName(reference.Branch)
		}

		repo, err = git.Clone(memory.NewStorage(), memfs.New(), &opts)
		if err != nil {
			// Return the original error
			return nil, cloneErr
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	if reference.Commit != "" {
		h, err := repo.ResolveRevision(plumbing.Revision(reference.Commit))
		if err != nil {
			return nil, err
		}

		if err := tree.Checkout(&git.CheckoutOptions{
			Hash: *h,
		}); err != nil {
			return nil, err
		}
	} else if reference.Tag != "" {
		if err := tree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(reference.Tag),
		}); err != nil {
			return nil, err
		}
	}

	f, err := tree.Filesystem.Open(reference.Path)
	if err != nil {
		return nil, err
	}

	return f, nil
}
