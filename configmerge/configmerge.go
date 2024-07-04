package configmerge

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/bitrise/models"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v2"
)

type repoInfo struct {
	defaultRemoteURL string
	branch           string
	commit           string
	tag              string
}

type Merger struct {
	repo     *git.Repository
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
		repo:     repo,
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
		Branch:     m.repoInfo.branch,
		Commit:     m.repoInfo.commit,
		Tag:        m.repoInfo.tag,
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
		reader, err := openConfigModule(include, m.repoInfo.defaultRemoteURL)
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

func openConfigModule(reference configReference, defaultRemoteURL string) (io.Reader, error) {
	if reference.Repository == "" || reference.Repository == defaultRemoteURL {
		return openLocalConfigModule(reference)
	} else {
		return openRemoteConfigModule(reference)
	}
}

func openLocalConfigModule(reference configReference) (io.Reader, error) {
	return os.Open(reference.Path)
}

func openRemoteConfigModule(reference configReference) (io.Reader, error) {
	return nil, nil
}

func readRepoInfo(repo *git.Repository) (*repoInfo, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	var branch string
	var commit string
	var tag string

	if head.Name().IsBranch() {
		branch = head.Name().Short()
	} else if head.Name().IsTag() {
		tag = head.Name().Short()
	} else {
		commit = head.Hash().String()
	}

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

	return &repoInfo{
		defaultRemoteURL: remoteURL,
		branch:           branch,
		commit:           commit,
		tag:              tag,
	}, nil
}
