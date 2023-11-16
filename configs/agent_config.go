package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"gopkg.in/yaml.v2"
)

const agentConfigFileName = "agent-config.yml"
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

func RegisterAgentOverrides() error {
	if !hasAgentConfigFile() {
		return nil
	}

	file := filepath.Join(GetBitriseHomeDirPath(), agentConfigFileName)

	log.Print("")
	log.Info("Running in agent mode")
	log.Printf("Config file: %s", file)

	config, err := readAgentConfig(file)
	if err != nil {
		return fmt.Errorf("agent config file: %w", err)
	}

	params := []struct {
		dir    string
		envKey string
	}{
		{
			dir:    config.BitriseDirs.BitriseDataHomeDir,
			envKey: BitriseDataHomeDirEnvKey,
		},
		{
			dir:    config.BitriseDirs.SourceDir,
			envKey: BitriseSourceDirEnvKey,
		},
		{
			dir:    config.BitriseDirs.DeployDir,
			envKey: BitriseDeployDirEnvKey,
		},
		{
			dir:    config.BitriseDirs.TestDeployDir,
			envKey: BitriseTestDeployDirEnvKey,
		},
	}
	for _, param := range params {
		err = pathutil.EnsureDirExist(param.dir)
		if err != nil {
			return fmt.Errorf("can't create %s: %w", param.envKey, err)
		}
		err = os.Setenv(param.envKey, param.dir)
		if err != nil {
			return fmt.Errorf("set %s: %w", param.envKey, err)
		}
	}

	return nil
}

func hasAgentConfigFile() bool {
	exists, _ := pathutil.IsPathExists(filepath.Join(GetBitriseHomeDirPath(), agentConfigFileName))
	return exists
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

	dataHomeDir, err := normalizePath(config.BitriseDirs.BitriseDataHomeDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_DATA_HOME_DIR value: %s", err)
	}
	config.BitriseDirs.BitriseDataHomeDir = dataHomeDir

	// BITRISE_SOURCE_DIR
	if config.BitriseDirs.SourceDir == "" {
		config.BitriseDirs.SourceDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultSourceDir)
	}
	sourceDir, err := normalizePath(config.BitriseDirs.SourceDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_SOURCE_DIR value: %s", err)
	}
	config.BitriseDirs.SourceDir = sourceDir

	// BITRISE_DEPLOY_DIR
	if config.BitriseDirs.DeployDir == "" {
		config.BitriseDirs.DeployDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultDeployDir)
	}
	deployDir, err := normalizePath(config.BitriseDirs.DeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.DeployDir = deployDir

	// BITRISE_TEST_DEPLOY_DIR
	if config.BitriseDirs.TestDeployDir == "" {
		config.BitriseDirs.TestDeployDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultTestDeployDir)
	}
	testDeployDir, err := normalizePath(config.BitriseDirs.TestDeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_TEST_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.TestDeployDir = testDeployDir

	// Hooks
	if config.Hooks.DoOnWorkflowStart != "" {
		doOnWorkflowStart, err := normalizePath(config.Hooks.DoOnWorkflowStart)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_workflow_start value: %s", err)
		}
		doOnWorkflowStartExists, err := pathutil.IsPathExists(doOnWorkflowStart)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnWorkflowStartExists {
			return AgentConfig{}, fmt.Errorf("do_on_workflow_start path does not exist: %s", doOnWorkflowStart)
		}
		config.Hooks.DoOnWorkflowStart = doOnWorkflowStart
	}

	if config.Hooks.DoOnWorkflowEnd != "" {
		doOnWorkflowEnd, err := normalizePath(config.Hooks.DoOnWorkflowEnd)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_workflow_end value: %s", err)
		}
		doOnWorkflowEndExists, err := pathutil.IsPathExists(doOnWorkflowEnd)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnWorkflowEndExists {
			return AgentConfig{}, fmt.Errorf("do_on_workflow_end path does not exist: %s", doOnWorkflowEnd)
		}
		config.Hooks.DoOnWorkflowEnd = doOnWorkflowEnd
	}

	return config, nil
}

func normalizePath(path string) (string, error) {
	expanded, err := pathutil.ExpandTilde(os.ExpandEnv(path))
	if err != nil {
		return "", err
	}
	return pathutil.AbsPath(expanded)
}
