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

type ConfigReader interface {
	Read(ref ConfigReference, dir string) ([]byte, error)
}

type Merger struct {
	configReader ConfigReader
	logger       logV2.Logger

	filesCount int
}

func NewMerger(configReader ConfigReader, logger logV2.Logger) Merger {
	return Merger{
		configReader: configReader,
		logger:       logger,
	}
}

func (m *Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	repoDir := filepath.Dir(mainConfigPth)
	mainConfigRef := ConfigReference{
		Path: mainConfigPth,
	}

	mainConfigBytes, err := m.configReader.Read(mainConfigRef, repoDir)
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
		moduleBytes, err := m.configReader.Read(include, dir)
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
