package integration

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/log"
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

func collectStepOutputs(workflowRunOut string, t *testing.T) []string {
	type messageLine struct {
		Producer   string `json:"producer"`
		ProducerID string `json:"producer_id"`
		Message    string `json:"message"`
	}

	idxToStep := map[string]int{}
	messageToStep := map[string]string{}
	stepIDx := 0

	for _, line := range strings.Split(workflowRunOut, "\n") {
		var message messageLine
		err := json.Unmarshal([]byte(line), &message)
		require.NoError(t, err)

		if message.Producer != "step" {
			continue
		}

		stepMessage := messageToStep[message.ProducerID]
		stepMessage += message.Message
		messageToStep[message.ProducerID] = stepMessage

		_, ok := idxToStep[message.ProducerID]
		if !ok {
			idxToStep[message.ProducerID] = stepIDx
			stepIDx++
		}
	}

	stepMessages := make([]string, len(idxToStep))
	for step, idx := range idxToStep {
		stepMessages[idx] = messageToStep[step]
	}

	return stepMessages
}
