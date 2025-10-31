package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
	"gopkg.in/yaml.v2"
)

const (
	agentConfigFileName  = "agent-config.yml"
	defaultSourceDir     = "workspace"
	defaultDeployDir     = "$BITRISE_APP_SLUG/$BITRISE_BUILD_SLUG/artifacts"
	defaultTestDeployDir = "$BITRISE_APP_SLUG/$BITRISE_BUILD_SLUG/test_results"
	defaultHTMLReportDir = "$BITRISE_APP_SLUG/$BITRISE_BUILD_SLUG/html_reports"
)

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

	// HTMLReportDir is for deployable html reports.
	// It might be outside of BitriseDataHomeDir if the user has configured it so
	HTMLReportDir string `yaml:"BITRISE_HTML_REPORT_DIR"`
}

// AgentHooks are various hooks that are executed before and after the CLI runs a build.
// In this context, a build means a CI execution triggered automatically (unlike a local invocation by `bitrise run x`).
// These hooks are only executed once, even if the actual workflow triggers other workflows.
type AgentHooks struct {
	// CleanupOnBuildStart is the list of UNEXPANDED paths to clean up before executing the main workflow.
	// The actual string value should be expanded at execution time, so that
	// Bitrise dirs defined in this config file are correctly expanded.
	CleanupOnBuildStart []string `yaml:"cleanup_on_build_start"`

	// CleanupOnBuildEnd is the list of UNEXPANDED paths to clean up after executing the main workflow.
	// The actual string value should be expanded at execution time, so that
	// Bitrise dirs defined in this config file are correctly expanded.
	CleanupOnBuildEnd []string `yaml:"cleanup_on_build_end"`

	// DoOnBuildStart is an optional executable to run before executing the main workflow..
	DoOnBuildStart string `yaml:"do_on_build_start"`

	// DoOnBuildEnd is an optional executable to run after executing the main workflow.
	DoOnBuildEnd string `yaml:"do_on_build_end"`
}

func GetAgentConfigPath() string {
	return filepath.Join(GetBitriseHomeDirPath(), agentConfigFileName)
}

func HasAgentConfig() bool {
	exists, _ := pathutil.IsPathExists(GetAgentConfigPath())
	return exists
}

func ReadAgentConfig(configFile string) (AgentConfig, error) {
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

	// BITRISE_HTML_REPORT_DIR
	if config.BitriseDirs.HTMLReportDir == "" {
		config.BitriseDirs.HTMLReportDir = filepath.Join(config.BitriseDirs.BitriseDataHomeDir, defaultHTMLReportDir)
	}
	htmlReportDir, err := normalizePath(config.BitriseDirs.HTMLReportDir)
	if err != nil {
		return AgentConfig{}, fmt.Errorf("expand BITRISE_HTML_REPORT_DIR value: %w", err)
	}
	config.BitriseDirs.HTMLReportDir = htmlReportDir

	// Hooks
	if config.Hooks.DoOnBuildStart != "" {
		doOnBuildStart, err := normalizePath(config.Hooks.DoOnBuildStart)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_build_start value: %s", err)
		}
		doOnBuildStartExists, err := pathutil.IsPathExists(doOnBuildStart)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnBuildStartExists {
			return AgentConfig{}, fmt.Errorf("do_on_build_start path does not exist: %s", doOnBuildStart)
		}
		config.Hooks.DoOnBuildStart = doOnBuildStart
	}

	if config.Hooks.DoOnBuildEnd != "" {
		doOnBuildEnd, err := normalizePath(config.Hooks.DoOnBuildEnd)
		if err != nil {
			return AgentConfig{}, fmt.Errorf("expand do_on_build_end value: %s", err)
		}
		doOnBuildEndExists, err := pathutil.IsPathExists(doOnBuildEnd)
		if err != nil {
			return AgentConfig{}, err
		}
		if !doOnBuildEndExists {
			return AgentConfig{}, fmt.Errorf("do_on_build_end path does not exist: %s", doOnBuildEnd)
		}
		config.Hooks.DoOnBuildEnd = doOnBuildEnd
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
