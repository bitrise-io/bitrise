package version

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/hashicorp/go-version"
)

// StepmanVersion ...
func StepmanVersion() (version.Version, error) {
	args := []string{"--version"}

	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return version.Version{}, fmt.Errorf("stepman --version: %s", versionOut)
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("parse version %s", versionOut)
	}

	return *versionPtr, nil
}

// EnvmanVersion ...
func EnvmanVersion(binPath string) (version.Version, error) {
	args := []string{"envman", "--version"}
	versionOut, err := command.RunCommandAndReturnCombinedStdoutAndStderr(binPath, args...)
	if err != nil {
		return version.Version{}, fmt.Errorf("envman --version: %s", versionOut)
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("parse version %s", versionOut)
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
		return version.Version{}, fmt.Errorf("parse version %s", VERSION)
	}

	return *versionPtr, nil
}

// ToolVersionMap ...
func ToolVersionMap(binPath string) (map[string]version.Version, error) {
	envmanVersion, err := EnvmanVersion(binPath)
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
