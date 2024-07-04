package configmerge

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/bitrise/models"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v2"
)

type configReference struct {
	Repository string `yaml:"repository" json:"repository"`
	Branch     string `yaml:"branch" json:"branch"`
	Commit     string `yaml:"commit" json:"commit"`
	Tag        string `yaml:"tag" json:"tag"`
	Path       string `yaml:"path" json:"path"`
}

func (r configReference) Key() string {
	var key string
	if r.Repository != "" {
		key = fmt.Sprintf("repo:%s,%s", r.Repository, r.Path)
	} else {
		key = r.Path
	}

	if r.Commit != "" {
		key += fmt.Sprintf("@%s", r.Commit)
	} else if r.Tag != "" {
		key += fmt.Sprintf("@%s", r.Tag)
	} else if r.Branch != "" {
		key += fmt.Sprintf("@%s", r.Branch)
	}

	return key
}

type configWithIncludes struct {
	Include []configReference `yaml:"include" json:"include"`
}

type Merger struct {
	repo             *git.Repository
	defaultRemoteURL string
}

func NewMerger(repoPth string) (*Merger, error) {
	repo, err := git.PlainOpen(repoPth)
	if err != nil {
		return nil, err
	}

	url, err := readDefaultRemoteURL(repo)
	if err != nil {
		return nil, err
	}

	return &Merger{
		repo:             repo,
		defaultRemoteURL: url,
	}, nil
}

func (m Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	f, err := os.Open(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	ref := configReference{
		Repository: m.defaultRemoteURL,
		Branch:     "",
		Commit:     "",
		Tag:        "",
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

	var config configWithIncludes
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	var includedConfigs []models.ConfigFileTreeModel
	for _, include := range config.Include {
		reader, err := openConfigModule(include, m.defaultRemoteURL)
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

func readDefaultRemoteURL(repo *git.Repository) (string, error) {
	config, err := repo.Config()
	if err != nil {
		return "", err
	}

	remote, ok := config.Remotes[git.DefaultRemoteName]
	if !ok {
		return "", fmt.Errorf("remote %s not exists", git.DefaultRemoteName)
	}

	if len(remote.URLs) == 0 {
		return "", fmt.Errorf("URL not set for remote %s", git.DefaultRemoteName)
	}

	return remote.URLs[0], nil
}
