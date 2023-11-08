package configs

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/pathutil"
	yaml "gopkg.in/yaml.v1"
)

const agentConfigFileName = "agent-config.yml"

type AgentConfig struct {
	BitriseDirs BitriseDirs `yaml:"bitrise_dirs"`
	Hooks       AgentHooks  `yaml:"hooks"`
}

type BitriseDirs struct {
	SourceDir     string `yaml:"BITRISE_SOURCE_DIR"`
	DeployDir     string `yaml:"BITRISE_DEPLOY_DIR"`
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

	DoOnWorkflowStart string `yaml:"do_on_workflow_start"`
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

	expandedSourceDir, err := expandPath(config.BitriseDirs.SourceDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_SOURCE_DIR value: %s", err)
	}
	config.BitriseDirs.SourceDir = expandedSourceDir

	expandedDeployDir, err := expandPath(config.BitriseDirs.DeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.DeployDir = expandedDeployDir

	expandedTestDeployDir, err := expandPath(config.BitriseDirs.TestDeployDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_TEST_DEPLOY_DIR value: %s", err)
	}
	config.BitriseDirs.TestDeployDir = expandedTestDeployDir

	expandedDoOnWorkflowStart, err := expandPath(config.Hooks.DoOnWorkflowStart)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand do_on_workflow_start value: %s", err)
	}
	config.Hooks.DoOnWorkflowStart = expandedDoOnWorkflowStart

	expandedDoOnWorkflowEnd, err := expandPath(config.Hooks.DoOnWorkflowEnd)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand do_on_workflow_end value: %s", err)
	}
	config.Hooks.DoOnWorkflowEnd = expandedDoOnWorkflowEnd

	return config, nil
}

func expandPath(path string) (string, error) {
	return pathutil.ExpandTilde(os.ExpandEnv(path))
}
