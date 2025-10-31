package utils

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CheckProgramInstalledPath ...
func CheckProgramInstalledPath(clcommand string) (string, error) {
	cmd := exec.Command("which", clcommand)
	outBytes, err := cmd.Output()
	outStr := string(outBytes)
	return strings.TrimSpace(outStr), err
}

// FormattedSecondsToMax8Chars ...
func FormattedSecondsToMax8Chars(t time.Duration) (string, error) {
	sec := t.Seconds()
	min := t.Minutes()
	hour := t.Hours()

	if sec < 1.0 {
		// 0.999999 sec -> 0.99 sec
		return fmt.Sprintf("%.2f sec", sec), nil // 8
	} else if sec < 60.0 {
		// 59.99999 sec -> 59.99 sec
		return fmt.Sprintf("%.2f sec", sec), nil // 8
	} else if min < 60 {
		// 59,999 min -> 59.9 min
		return fmt.Sprintf("%.1f min", min), nil // 8
	} else if hour < 10 {
		// 9.999 hour -> 9.9 hour
		return fmt.Sprintf("%.1f hour", hour), nil // 8
	} else if hour < 1000 {
		// 999,999 hour -> 999 hour
		return fmt.Sprintf("%.f hour", hour), nil // 8
	}

	return "", fmt.Errorf("time (%f hour) greater than max allowed (999 hour)", hour)
}
