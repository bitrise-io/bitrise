package utils

import (
	"fmt"
	"runtime"
)

// UnameGOOS ...
func UnameGOOS() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin", nil
	case "linux":
		return "Linux", nil
	}
	return "", fmt.Errorf("Unsupported platform (%s)", runtime.GOOS)
}

// UnameGOARCH ...
func UnameGOARCH() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", nil
	case "arm64":
		return "arm64", nil
	}
	return "", fmt.Errorf("Unsupported architecture (%s)", runtime.GOARCH)
}
