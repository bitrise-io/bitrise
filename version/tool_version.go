package version

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/hashicorp/go-version"
)

// StepmanVersion ...
func StepmanVersion(binPath string) (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"stepman", "--loglevel", logLevel, "--version"}

	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, args...)
	if err != nil {
		return version.Version{}, err
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("Failed to parse version (%s)", versionOut)
	}

	return *versionPtr, nil
}

// EnvmanVersion ...
func EnvmanVersion(binPath string) (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"envman", "--loglevel", logLevel, "--version"}
	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, args...)
	if err != nil {
		return version.Version{}, err
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("Failed to parse version (%s)", versionOut)
	}

	return *versionPtr, nil
}

// BitriseCliVersion ...
func BitriseCliVersion() (version.Version, error) {
	versionPtr, err := version.NewVersion(VERSION)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("Failed to parse version (%s)", VERSION)
	}

	return *versionPtr, nil
}

// ToolVersionMap ...
func ToolVersionMap(binPath string) (map[string]version.Version, error) {
	envmanVersion, err := EnvmanVersion(binPath)
	if err != nil {
		return map[string]version.Version{}, err
	}

	stepmanVersion, err := StepmanVersion(binPath)
	if err != nil {
		return map[string]version.Version{}, err
	}

	bitriseVersion, err := BitriseCliVersion()
	if err != nil {
		return map[string]version.Version{}, err
	}

	return map[string]version.Version{
		"bitrise": bitriseVersion,
		"envman":  envmanVersion,
		"stepman": stepmanVersion,
	}, nil
}
