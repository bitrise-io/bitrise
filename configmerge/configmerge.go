package configmerge

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/sliceutil"
	"gopkg.in/yaml.v2"
)

const (
	MaxIncludeCountPerFile = 10
	MaxFilesCountTotal     = 20
	MaxIncludeDepth        = 5           // root + 4 includes
	MaxFileSizeBytes       = 1024 * 1024 // 1MB
)

type ConfigModule struct {
	Include []ConfigReference `yaml:"include" json:"include"`
}

func IsModularConfig(mainConfigPth string) (bool, error) {
	mainConfigFile, err := os.Open(mainConfigPth)
	if err != nil {
		return false, err
	}
	mainConfigContent, err := io.ReadAll(mainConfigFile)
	if err != nil {
		return false, err
	}

	var config ConfigModule
	if err := yaml.Unmarshal(mainConfigContent, &config); err != nil {
		return false, err
	}
	return len(config.Include) > 0, nil
}

type ConfigReader interface {
	Read(ref ConfigReference) ([]byte, error)
	CleanupRepoDirs() error
}

type Merger struct {
	configReader ConfigReader
	logger       log.Logger

	filesCount int
}

func NewMerger(configReader ConfigReader, logger log.Logger) Merger {
	return Merger{
		configReader: configReader,
		logger:       logger,
	}
}

func (m *Merger) MergeConfig(mainConfigPth string) (string, *models.ConfigFileTreeModel, error) {
	defer func() {
		if err := m.configReader.CleanupRepoDirs(); err != nil {
			m.logger.Warnf("Failed to cleanup modular config local cache dir: %s", err)
		}
	}()

	m.logger.Debugf("Merge config modules included in %s", mainConfigPth)

	mainConfigRef := ConfigReference{
		Path: mainConfigPth,
	}

	mainConfigBytes, err := m.configReader.Read(mainConfigRef)
	if err != nil {
		return "", nil, err
	}

	m.logger.Debugf("Building config tree")

	configTree, err := m.buildConfigTree(mainConfigBytes, mainConfigRef, 1, nil)
	if err != nil {
		return "", nil, err
	}

	m.logger.Debugf("Merging config tree")

	mergedConfigContent, err := configTree.Merge()
	if err != nil {
		return "", nil, err
	}

	return mergedConfigContent, configTree, nil
}

func (m *Merger) buildConfigTree(configContent []byte, reference ConfigReference, depth int, keys []string) (*models.ConfigFileTreeModel, error) {
	key := reference.Key()
	keys = append(keys, key)

	m.filesCount++

	var config ConfigModule
	if err := yaml.Unmarshal(configContent, &config); err != nil {
		return nil, err
	}

	for idx, include := range config.Include {
		if include.Repository == "" {
			include.Repository = reference.Repository
			include.Branch = reference.Branch
			include.Commit = reference.Commit
			include.Tag = reference.Tag
		}

		config.Include[idx] = include
	}

	if err := validateReference(reference, configContent, config, m.filesCount, depth, keys); err != nil {
		return nil, err
	}

	var includedConfigTrees []models.ConfigFileTreeModel
	for _, include := range config.Include {
		moduleBytes, err := m.configReader.Read(include)
		if err != nil {
			return nil, err
		}

		moduleConfigTree, err := m.buildConfigTree(moduleBytes, include, depth+1, keys)
		if err != nil {
			return nil, err
		}

		includedConfigTrees = append(includedConfigTrees, *moduleConfigTree)
	}

	return &models.ConfigFileTreeModel{
		Path:     key,
		Contents: string(configContent),
		Includes: includedConfigTrees,
	}, nil
}

func validateReference(reference ConfigReference, configContent []byte, config ConfigModule, filesCount int, depth int, keys []string) error {
	key := reference.Key()

	if len(configContent) > MaxFileSizeBytes {
		return fmt.Errorf("max file size (%d bytes) exceeded in file %s", MaxFileSizeBytes, key)
	}

	if depth > MaxIncludeDepth {
		return fmt.Errorf("max include depth (%d) exceeded", MaxIncludeDepth)
	}

	if filesCount > MaxFilesCountTotal {
		return fmt.Errorf("max include count (%d) exceeded", MaxFilesCountTotal)
	}

	if len(config.Include) > MaxIncludeCountPerFile {
		return fmt.Errorf("max include count (%d) exceeded", MaxIncludeCountPerFile)
	}
	if filesCount+len(config.Include) > MaxFilesCountTotal {
		return fmt.Errorf("max file count (%d) exceeded", MaxFilesCountTotal)
	}

	for _, include := range config.Include {
		if err := include.Validate(); err != nil {
			return err
		}

		if sliceutil.IsStringInSlice(include.Key(), keys) {
			return fmt.Errorf("circular reference detected: %s -> %s", strings.Join(keys, " -> "), include.Key())
		}
	}

	return nil
}
