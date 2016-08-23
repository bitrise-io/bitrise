package errorutil

import (
	"errors"
	"os/exec"
	"regexp"
	"syscall"
)

// IsExitStatusError ...
func IsExitStatusError(err error) bool {
	return IsExitStatusErrorStr(err.Error())
}

// IsExitStatusErrorStr ...
func IsExitStatusErrorStr(errString string) bool {
	// example exit status error string: exit status 1
	var rex = regexp.MustCompile(`^exit status [0-9]{1,3}$`)
	return rex.MatchString(errString)
}

// CmdExitCodeFromError ...
func CmdExitCodeFromError(err error) (int, error) {
	cmdExitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return 1, errors.New("Failed to cast exit status")
			}
			cmdExitCode = waitStatus.ExitStatus()
		}
		return cmdExitCode, nil
	}
	return 0, nil
}
