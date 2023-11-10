package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
	"gopkg.in/yaml.v2"
)

const defaultSourceDir = "workspace"
const defaultDeployDir = "$BITRISE_APP_SLUG/$BITRISE_BUILD_SLUG/artifacts"
const defaultTestDeployDir = "$BITRISE_APP_SLUG/$BITRISE_BUILD_SLUG/test_results"

type AgentConfig struct {
	BitriseDirs BitriseDirs `yaml:"bitrise_dirs"`
	Hooks       AgentHooks  `yaml:"hooks"`
}

type BitriseDirs struct {
	// BitriseDataHomeDir is the root directory for all Bitrise data produced at runtime
	BitriseDataHomeDir string `yaml:"BITRISE_DATA_HOME_DIR"`

	// SourceDir is for source code checkouts.
	// It might be outside of BitriseDataHomeDir if the user has configured it so
	SourceDir string `yaml:"BITRISE_SOURCE_DIR"`

	// DeployDir is for deployable artifacts.
	// It might be outside of BitriseDataHomeDir if the user has configured it so
	DeployDir string `yaml:"BITRISE_DEPLOY_DIR"`

	// TestDeployDir is for deployable test result artifacts.
	// It might be outside of BitriseDataHomeDir if the user has configured it so
	TestDeployDir string `yaml:"BITRISE_TEST_DEPLOY_DIR"`
}

type AgentHooks struct {
	// CleanupOnWorkflowStart is the list of UNEXPANDED paths to clean up when the workflow starts.
	// The actual string value should be expanded at execution time, so that
	// Bitrise dirs defined in this config file are correctly expanded.
	CleanupOnWorkflowStart []string `yaml:"cleanup_on_workflow_start"`

	// CleanupOnWorkflowEnd is the list of UNEXPANDED paths to clean up when the workflow end.
	// The actual string value should be expanded at execution time, so that
	// Bitrise dirs defined in this config file are correctly expanded.
	CleanupOnWorkflowEnd []string `yaml:"cleanup_on_workflow_end"`

	// DoOnWorkflowStart is an optional executable to run when the workflow starts.
	DoOnWorkflowStart string `yaml:"do_on_workflow_start"`

	// DoOnWorkflowEnd is an optional executable to run when the workflow ends.
	DoOnWorkflowEnd   string `yaml:"do_on_workflow_end"`
}

func readAgentConfig(configFile string) (AgentConfig, error) {
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		return AgentConfig{}, err
	}

	var config AgentConfig
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return AgentConfig{}, err
	}

	expandedBitriseDataHomeDir, err := expandPath(config.BitriseDirs.BitriseDataHomeDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_DATA_HOME_DIR value: %s", err)
	}
	config.BitriseDirs.BitriseDataHomeDir = expandedBitriseDataHomeDir

	// BITRISE_SOURCE_DIR
	if config.BitriseDirs.SourceDir == "" {
		config.BitriseDirs.SourceDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultSourceDir)
	}
	expandedSourceDir, err := expandPath(config.BitriseDirs.SourceDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_SOURCE_DIR value: %s", err)
	}
	config.BitriseDirs.SourceDir = expandedSourceDir

	// BITRISE_DEPLOY_DIR
	if config.BitriseDirs.DeployDir == "" {
		config.BitriseDirs.DeployDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultDeployDir)
	}
	expandedDeployDir, err := expandPath(config.BitriseDirs.DeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.DeployDir = expandedDeployDir

	// BITRISE_TEST_DEPLOY_DIR
	if config.BitriseDirs.TestDeployDir == "" {
		config.BitriseDirs.TestDeployDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultTestDeployDir)
	}
	expandedTestDeployDir, err := expandPath(config.BitriseDirs.TestDeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_TEST_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.TestDeployDir = expandedTestDeployDir

	// Hooks
	if config.Hooks.DoOnWorkflowStart != "" {
		expandedDoOnWorkflowStart, err := expandPath(config.Hooks.DoOnWorkflowStart)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_workflow_start value: %s", err)
		}
		doOnWorkflowStartExists, err := pathutil.IsPathExists(expandedDoOnWorkflowStart)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnWorkflowStartExists {
			return AgentConfig{}, fmt.Errorf("do_on_workflow_start path does not exist: %s", expandedDoOnWorkflowStart)
		}
		config.Hooks.DoOnWorkflowStart = expandedDoOnWorkflowStart
	}

	if config.Hooks.DoOnWorkflowEnd != "" {
		expandedDoOnWorkflowEnd, err := expandPath(config.Hooks.DoOnWorkflowEnd)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_workflow_end value: %s", err)
		}
		doOnWorkflowEndExists, err := pathutil.IsPathExists(expandedDoOnWorkflowEnd)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnWorkflowEndExists {
			return AgentConfig{}, fmt.Errorf("do_on_workflow_end path does not exist: %s", expandedDoOnWorkflowEnd)
		}
		config.Hooks.DoOnWorkflowEnd = expandedDoOnWorkflowEnd
	}

	return config, nil
}

func expandPath(path string) (string, error) {
	return pathutil.ExpandTilde(os.ExpandEnv(path))
}
