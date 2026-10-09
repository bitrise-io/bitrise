package integration

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var binPathStr string

func binPath() string {
	if binPathStr != "" {
		return binPathStr
	}

	pth := os.Getenv("INTEGRATION_TEST_BINARY_PATH")
	if pth == "" {
		if os.Getenv("CI") == "true" {
			panic("INTEGRATION_TEST_BINARY_PATH env is required in CI")
		} else {
			log.Warn("INTEGRATION_TEST_BINARY_PATH is not set, make sure 'envman' binary in your PATH is up-to-date")
			pth = "envman"
		}
	}
	binPathStr = pth
	return pth
}
