package configmerge

import (
	"path/filepath"

	"github.com/bitrise-io/bitrise/models"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"gopkg.in/yaml.v2"
)

type RepoInfoProvider interface {
	GetRepoInfo(repoPth string) (*RepoInfo, error)
}

type FileReader interface {
	ReadFileFromFileSystem(name string) ([]byte, error)
	ReadFileFromGitRepository(repository string, branch string, commit string, tag string, path string) ([]byte, error)
}

type FileCache interface {
	GetFileContent(key string) ([]byte, error)
	SetFileContent(key string, content []byte) error
}

type Merger struct {
	repoInfoProvider RepoInfoProvider
	fileReader       FileReader
	fileCache        FileCache
	logger           logV2.Logger

	repoInfo RepoInfo
}

func NewMerger(repoInfoProvider RepoInfoProvider, fileReader FileReader, fileCache FileCache, logger logV2.Logger) Merger {
	return Merger{
		repoInfoProvider: repoInfoProvider,
		fileReader:       fileReader,
		fileCache:        fileCache,
		logger:           logger,
	}
}

func (m *Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	repoDir := filepath.Dir(mainConfigPth)
	info, err := m.repoInfoProvider.GetRepoInfo(repoDir)
	if err != nil {
		return "", nil, err
	}
	m.repoInfo = *info

	ref := ConfigReference{
		Repository: info.DefaultRemoteURL,
		Commit:     info.Commit,
		Tag:        info.Tag,
		Branch:     info.Branch,
		Path:       mainConfigPth,
	}

	b, err := m.fileReader.ReadFileFromFileSystem(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	configTree, err := m.buildConfigTree(b, ref)
	if err != nil {
		return "", nil, err
	}

	mergedConfigContent, err := configTree.Merge()
	if err != nil {
		return "", nil, err
	}

	return mergedConfigContent, configTree, nil
}

func (m *Merger) buildConfigTree(configContent []byte, reference ConfigReference) (*models.ConfigFileTreeModel, error) {
	var config struct {
		Include []ConfigReference `yaml:"include" json:"include"`
	}
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		return nil, err
	}

	var includedConfigs []models.ConfigFileTreeModel
	for _, include := range config.Include {
		b, err := m.readConfigModule(include, m.repoInfo)
		if err != nil {
			return nil, err
		}

		tree, err := m.buildConfigTree(b, include)
		if err != nil {
			return nil, err
		}

		includedConfigs = append(includedConfigs, *tree)
	}

	return &models.ConfigFileTreeModel{
		Path:     reference.Key(),
		Contents: string(configContent),
		Includes: includedConfigs,
	}, nil
}

func (m *Merger) readConfigModule(reference ConfigReference, info RepoInfo) ([]byte, error) {
	localReference, err := isLocalReference(reference, info)
	if err != nil {
		return nil, err
	}

	if localReference {
		return m.readLocalConfigModule(reference)
	} else {
		if m.fileCache != nil {
			b, err := m.fileCache.GetFileContent(reference.Key())
			if err != nil {
				m.logger.Warnf("Failed to read file (%s) from cache: %s", err)
			} else if len(b) > 0 {
				return b, nil
			}
		}

		b, err := m.readRemoteConfigModule(reference)
		if err != nil {
			return nil, err
		}

		if m.fileCache != nil {
			if err := m.fileCache.SetFileContent(reference.Key(), b); err != nil {
				m.logger.Warnf("Failed to cache file (%s): %s", err)
			}
		}

		return b, nil
	}
}

func isLocalReference(reference ConfigReference, info RepoInfo) (bool, error) {
	if reference.Repository == "" {
		return true, nil
	}

	refGitUrl, err := parseGitRepoURL(reference.Repository)
	if err != nil {
		return false, err
	}

	repoGitURL, err := parseGitRepoURL(info.DefaultRemoteURL)
	if err != nil {
		return false, err
	}

	if !equalGitRepoURLs(refGitUrl, repoGitURL) {
		return false, nil
	}

	switch {
	case reference.Commit != "":
		return reference.Commit == info.Commit ||
			reference.Commit == info.Commit[:7], nil
	case reference.Tag != "":
		return reference.Tag == info.Tag, nil
	case reference.Branch != "":
		return reference.Branch == info.Branch, nil
	}

	return true, nil
}

func (m *Merger) readLocalConfigModule(reference ConfigReference) ([]byte, error) {
	return m.fileReader.ReadFileFromFileSystem(reference.Path)
}

func (m *Merger) readRemoteConfigModule(reference ConfigReference) ([]byte, error) {
	return m.fileReader.ReadFileFromGitRepository(reference.Repository, reference.Branch, reference.Commit, reference.Tag, reference.Path)

}
