package errorutil

import "regexp"

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
