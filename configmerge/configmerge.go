package configmerge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/sliceutil"
	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"gopkg.in/yaml.v2"
)

const (
	MaxIncludeCountPerFile = 10
	MaxFilesCountTotal     = 20
	MaxIncludeDepth        = 5           // root + 4 includes
	MaxFileSizeBytes       = 1024 * 1024 // 1MB
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

	repoInfo *RepoInfo

	filesCount int
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
		m.logger.Debugf("Failed to get repository info: %s", err)
	} else {
		m.repoInfo = repoInfo
	}

	mainConfigRef := ConfigReference{
		Path: mainConfigPth,
	}

	if repoInfo != nil {
		mainConfigRef.Repository = repoInfo.DefaultRemoteURL
		mainConfigRef.Commit = repoInfo.Commit
		mainConfigRef.Tag = repoInfo.Tag
		mainConfigRef.Branch = repoInfo.Branch
	}

	mainConfigBytes, err := m.fileReader.ReadFileFromFileSystem(mainConfigPth)
	if err != nil {
		return "", nil, err
	}

	mainConfigDir := filepath.Dir(mainConfigPth)
	configTree, err := m.buildConfigTree(mainConfigBytes, mainConfigRef, mainConfigDir, 1, nil)
	if err != nil {
		return "", nil, err
	}

	mergedConfigContent, err := configTree.Merge()
	if err != nil {
		return "", nil, err
	}

	return mergedConfigContent, configTree, nil
}

func (m *Merger) buildConfigTree(configContent []byte, reference ConfigReference, dir string, depth int, keys []string) (*models.ConfigFileTreeModel, error) {
	key := reference.Key()
	keys = append(keys, key)

	if len(configContent) > MaxFileSizeBytes {
		return nil, fmt.Errorf("max file size (%d bytes) exceeded in file %s", MaxFileSizeBytes, key)
	}

	if depth > MaxIncludeDepth {
		return nil, fmt.Errorf("max include depth (%d) exceeded", MaxIncludeDepth)
	}

	m.filesCount++
	if m.filesCount > MaxFilesCountTotal {
		return nil, fmt.Errorf("max include count (%d) exceeded", MaxFilesCountTotal)
	}

	var config struct {
		Include []ConfigReference `yaml:"include" json:"include"`
	}
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		return nil, err
	}

	if len(config.Include) > MaxIncludeCountPerFile {
		return nil, fmt.Errorf("max include count (%d) exceeded", MaxIncludeCountPerFile)
	}
	if m.filesCount+len(config.Include) > MaxFilesCountTotal {
		return nil, fmt.Errorf("max file count (%d) exceeded", MaxFilesCountTotal)
	}

	for idx, include := range config.Include {
		if err := include.Validate(); err != nil {
			return nil, err
		}
		if include.Repository == "" {
			include.Repository = reference.Repository
			include.Branch = reference.Branch
			include.Commit = reference.Commit
			include.Tag = reference.Tag
		}
		config.Include[idx] = include

		if sliceutil.IsStringInSlice(include.Key(), keys) {
			return nil, fmt.Errorf("circular reference detected: %s -> %s", strings.Join(keys, " -> "), include.Key())
		}
	}

	var includedConfigTrees []models.ConfigFileTreeModel
	for _, include := range config.Include {
		moduleBytes, err := m.readConfigModule(include, dir, m.repoInfo)
		if err != nil {
			return nil, err
		}

		moduleDir := filepath.Dir(include.Path)
		moduleConfigTree, err := m.buildConfigTree(moduleBytes, include, moduleDir, depth+1, keys)
		if err != nil {
			return nil, err
		}

		includedConfigTrees = append(includedConfigTrees, *moduleConfigTree)
	}

	return &models.ConfigFileTreeModel{
		Path:     key,
		Contents: string(configContent),
		Includes: includedConfigTrees,
		Depth:    depth,
	}, nil
}

func (m *Merger) readConfigModule(reference ConfigReference, dir string, repoInfo *RepoInfo) ([]byte, error) {
	if isLocalReference(reference) {
		return m.readLocalConfigModule(reference, dir)
	}

	sameRepo := false
	if repoInfo != nil {
		var err error
		if sameRepo, err = isSameRepoReference(reference, *repoInfo); err != nil {
			m.logger.Warnf("Failed to check if the reference is from the same repository: %s", err)
		}
	}

	if sameRepo {
		return m.readLocalConfigModule(reference, dir)
	}

	return m.readRemoteConfigModule(reference)
}

func isSameRepoReference(reference ConfigReference, repoInfo RepoInfo) (bool, error) {
	refGitUrl, err := parseGitRepoURL(reference.Repository)
	if err != nil {
		return false, err
	}

	repoGitURL, err := parseGitRepoURL(repoInfo.DefaultRemoteURL)
	if err != nil {
		return false, err
	}

	if !equalGitRepoURLs(refGitUrl, repoGitURL) {
		return false, nil
	}

	switch {
	case reference.Commit != "":
		return reference.Commit == repoInfo.Commit ||
			reference.Commit == repoInfo.Commit[:7], nil
	case reference.Tag != "":
		return reference.Tag == repoInfo.Tag, nil
	case reference.Branch != "":
		return reference.Branch == repoInfo.Branch, nil
	}

	return true, nil
}

func isLocalReference(reference ConfigReference) bool {
	return reference.Repository == ""
}

func (m *Merger) readLocalConfigModule(reference ConfigReference, dir string) ([]byte, error) {
	pth := reference.Path
	if !filepath.IsAbs(pth) {
		pth = filepath.Join(dir, pth)
	}
	return m.fileReader.ReadFileFromFileSystem(pth)
}

func (m *Merger) readRemoteConfigModule(reference ConfigReference) ([]byte, error) {
	return m.fileReader.ReadFileFromGitRepository(reference.Repository, reference.Branch, reference.Commit, reference.Tag, reference.Path)

}
