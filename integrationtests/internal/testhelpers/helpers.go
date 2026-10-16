package testhelpers

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/stretchr/testify/require"
)

var binPathStr string

func BinPath() string {
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

// IsolatedHome creates a HOME for a bitrise process in which ~/.bitrise and ~/.stepman
// start out empty. A test needs it when it changes state that other tests read from those
// dirs, such as the steplib cache or the agent config, because the packages run in parallel.
// Every other entry links to the real HOME, so the tools that steps run keep working: on the
// stacks, asdf, rbenv and mise keep their installs and global versions in HOME.
func IsolatedHome(t *testing.T) string {
	realHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	entries, err := os.ReadDir(realHomeDir)
	require.NoError(t, err)

	homeDir := t.TempDir()
	for _, entry := range entries {
		if entry.Name() == ".bitrise" || entry.Name() == ".stepman" {
			continue
		}
		err := os.Symlink(filepath.Join(realHomeDir, entry.Name()), filepath.Join(homeDir, entry.Name()))
		require.NoError(t, err)
	}

	return homeDir
}

func ToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func ToJSON(t *testing.T, stringStringMap map[string]interface{}) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func WithRunningTimeCheck(f func()) time.Duration {
	start := time.Now()
	f()
	end := time.Now()

	return end.Sub(start)
}

func CollectStepOutputs(workflowRunOut string, t *testing.T) []string {
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