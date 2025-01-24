package integration

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"
	"github.com/stretchr/testify/require"
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
			log.Warn("INTEGRATION_TEST_BINARY_PATH is not set, make sure 'bitrise' binary in your PATH is up-to-date")
			pth = "bitrise"
		}
	}
	binPathStr = pth
	return pth
}

func toBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func toJSON(t *testing.T, stringStringMap map[string]interface{}) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func withRunningTimeCheck(f func()) time.Duration {
	start := time.Now()
	f()
	end := time.Now()

	return end.Sub(start)
}
