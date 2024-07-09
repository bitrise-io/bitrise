package configmerge

import (
	"io"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/models"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"gopkg.in/yaml.v2"
)

func IsModularConfig(mainConfigPth string) (bool, error) {
	mainConfigFile, err := os.Open(mainConfigPth)
	if err != nil {
		return false, err
	}
	mainConfigContent, err := io.ReadAll(mainConfigFile)
	if err != nil {
		return false, err
	}

	var config struct {
		Include []ConfigReference `yaml:"include" json:"include"`
	}
	if err := yaml.Unmarshal(mainConfigContent, &config); err != nil {
		return false, err
	}
	return len(config.Include) > 0, nil
}

type RepoInfoProvider interface {
	GetRepoInfo(repoPth string) (*RepoInfo, error)
}

type FileReader interface {
	ReadFileFromFileSystem(name string) ([]byte, error)
	ReadFileFromGitRepository(repository string, branch string, commit string, tag string, path string) ([]byte, error)
}

type Merger struct {
	repoInfoProvider RepoInfoProvider
	fileReader       FileReader
	logger           logV2.Logger

	repoInfo RepoInfo
}

func NewMerger(repoInfoProvider RepoInfoProvider, fileReader FileReader, logger logV2.Logger) Merger {
	return Merger{
		repoInfoProvider: repoInfoProvider,
		fileReader:       fileReader,
		logger:           logger,
	}
}

func (m *Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	repoDir := filepath.Dir(mainConfigPth)
	repoInfo, err := m.repoInfoProvider.GetRepoInfo(repoDir)
	if err != nil {
		return "", nil, err
	}
	m.repoInfo = *repoInfo

	mainConfigRef := ConfigReference{
		Repository: repoInfo.DefaultRemoteURL,
		Commit:     repoInfo.Commit,
		Tag:        repoInfo.Tag,
		Branch:     repoInfo.Branch,
		Path:       mainConfigPth,
	}

	mainConfigBytes, err := m.fileReader.ReadFileFromFileSystem(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	configTree, err := m.buildConfigTree(mainConfigBytes, mainConfigRef)
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
	for _, include := range config.Include {
		if err := include.Validate(); err != nil {
			return nil, err
		}
	}

	var includedConfigTrees []models.ConfigFileTreeModel
	for _, include := range config.Include {
		moduleBytes, err := m.readConfigModule(include, m.repoInfo)
		if err != nil {
			return nil, err
		}

		moduleConfigTree, err := m.buildConfigTree(moduleBytes, include)
		if err != nil {
			return nil, err
		}

		includedConfigTrees = append(includedConfigTrees, *moduleConfigTree)
	}

	return &models.ConfigFileTreeModel{
		Path:     reference.Key(),
		Contents: string(configContent),
		Includes: includedConfigTrees,
	}, nil
}

func (m *Merger) readConfigModule(reference ConfigReference, info RepoInfo) ([]byte, error) {
	if isLocalReference(reference) {
		return m.readLocalConfigModule(reference)
	}

	if sameRepo, err := isSameRepoReference(reference, info); err != nil {
		m.logger.Warnf("Failed to check if the reference is from the same repository: %s", err)
	} else if sameRepo {
		return m.readLocalConfigModule(reference)
	}

	return m.readRemoteConfigModule(reference)
}

func isSameRepoReference(reference ConfigReference, info RepoInfo) (bool, error) {
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

func isLocalReference(reference ConfigReference) bool {
	return reference.Repository == ""
}

func (m *Merger) readLocalConfigModule(reference ConfigReference) ([]byte, error) {
	return m.fileReader.ReadFileFromFileSystem(reference.Path)
}

func (m *Merger) readRemoteConfigModule(reference ConfigReference) ([]byte, error) {
	return m.fileReader.ReadFileFromGitRepository(reference.Repository, reference.Branch, reference.Commit, reference.Tag, reference.Path)

}
