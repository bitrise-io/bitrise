package testhelpers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/stretchr/testify/require"
)

var BinPath = sync.OnceValue(func() string {
	pth := os.Getenv("INTEGRATION_TEST_BINARY_PATH")
	if pth == "" {
		if os.Getenv("CI") == "true" {
			panic("INTEGRATION_TEST_BINARY_PATH env is required in CI")
		} else {
			log.Warn("INTEGRATION_TEST_BINARY_PATH is not set, make sure 'bitrise' binary in your PATH is up-to-date")
			pth = "bitrise"
		}
	}
	return pth
})

// IsolatedHome creates a HOME for a bitrise process in which ~/.bitrise, ~/.stepman and
// ~/.config/bitrise start out empty. A test needs it when it changes state that other tests
// read from those dirs, such as the steplib cache, the Step binary cache, the agent config
// or the record of the version set up, because the packages run in parallel.
// Every other entry links to the real HOME, so the tools that steps run keep working: on the
// stacks, asdf, rbenv and mise keep their installs and global versions in HOME and
// ~/.config.
// ~/.bitrise/tools is the exception: it links to the real one when that exists. With no
// record of the version set up, the first bitrise run sets up, and finds envman and stepman
// there instead of downloading them.
func IsolatedHome(t *testing.T) string {
	realHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	homeDir := t.TempDir()
	linkEntries(t, realHomeDir, homeDir, ".bitrise", ".stepman", ".config")
	linkEntries(t, filepath.Join(realHomeDir, ".config"), filepath.Join(homeDir, ".config"), "bitrise")

	bitriseHomeDir := filepath.Join(homeDir, ".bitrise")
	err = os.Mkdir(bitriseHomeDir, 0755)
	require.NoError(t, err)

	// Without setup there is nothing to link, and setup can't install through a dangling link.
	realToolsDir := filepath.Join(realHomeDir, ".bitrise", "tools")
	if _, err := os.Stat(realToolsDir); err == nil {
		err := os.Symlink(realToolsDir, filepath.Join(bitriseHomeDir, "tools"))
		require.NoError(t, err)
	}

	return homeDir
}

// linkEntries creates dstDir and links every entry of srcDir into it, except the skipped
// ones. Without srcDir there is nothing to link, and dstDir starts out empty.
func linkEntries(t *testing.T, srcDir, dstDir string, skip ...string) {
	err := os.MkdirAll(dstDir, 0755)
	require.NoError(t, err)

	entries, err := os.ReadDir(srcDir)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	require.NoError(t, err)

	for _, entry := range entries {
		if slices.Contains(skip, entry.Name()) {
			continue
		}
		err := os.Symlink(filepath.Join(srcDir, entry.Name()), filepath.Join(dstDir, entry.Name()))
		require.NoError(t, err)
	}
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
