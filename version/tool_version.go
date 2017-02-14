package version

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/command"
	"github.com/hashicorp/go-version"
)

// StepmanVersion ...
func StepmanVersion() (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--version"}

	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
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
func EnvmanVersion() (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--version"}
	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr("envman", args...)
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
func ToolVersionMap() (map[string]version.Version, error) {
	envmanVersion, err := EnvmanVersion()
	if err != nil {
		return map[string]version.Version{}, err
	}

	stepmanVersion, err := StepmanVersion()
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
