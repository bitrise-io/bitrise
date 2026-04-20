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

	bitriseVersionPtr, err := version.NewVersion(VERSION)
	if err != nil {
		// Dev builds (no ldflags) have VERSION="dev" which is not valid semver.
		// Use a high sentinel so all plugin min-version requirements are satisfied.
		bitriseVersionPtr = version.Must(version.NewVersion("99.99.99"))
	}

	return map[string]version.Version{
		"bitrise": *bitriseVersionPtr,
		"envman":  envmanVersion,
		"stepman": stepmanVersion,
	}, nil
}
