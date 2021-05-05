// Package errorutil ...
package errorutil

import (
	"os/exec"
	"regexp"
)

func exitCode(err error) int {
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ProcessState.ExitCode()
	}
	return -1
}

// IsExitStatusError ...
func IsExitStatusError(err error) bool {
	return exitCode(err) != -1
}

// IsExitStatusErrorStr ...
func IsExitStatusErrorStr(errString string) bool {
	// https://golang.org/src/os/exec_posix.go?s=2421:2459#L87
	// example exit status error string: exit status 1
	var rex = regexp.MustCompile(`^exit status [0-9]{1,3}$`)
	return rex.MatchString(errString)
}

// CmdExitCodeFromError ...
func CmdExitCodeFromError(err error) (int, error) {
	return exitCode(err), err
}
