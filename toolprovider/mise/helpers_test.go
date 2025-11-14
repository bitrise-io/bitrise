package mise

import (
	"fmt"
	"strings"
	"time"
)

type fakeExecEnv struct {
	// responses maps command strings to their outputs
	responses map[string]string
	// errors maps command strings to errors
	errors map[string]error
}

func newFakeExecEnv() *fakeExecEnv {
	return &fakeExecEnv{
		responses: make(map[string]string),
		errors:    make(map[string]error),
	}
}

func (m *fakeExecEnv) setResponse(cmdKey string, output string) {
	m.responses[cmdKey] = output
}

func (m *fakeExecEnv) setError(cmdKey string, err error) {
	m.errors[cmdKey] = err
}

func (m *fakeExecEnv) InstallDir() string {
	return "/fake/mise/install/dir"
}

func (m *fakeExecEnv) RunMise(args ...string) (string, error) {
	return m.runCommand(args...)
}

func (m *fakeExecEnv) RunMisePlugin(args ...string) (string, error) {
	return m.runCommand(append([]string{"plugin"}, args...)...)
}

func (m *fakeExecEnv) RunMiseWithTimeout(timeout time.Duration, args ...string) (string, error) {
	return m.runCommand(args...)
}

func (m *fakeExecEnv) runCommand(args ...string) (string, error) {
	cmdKey := strings.Join(args, " ")

	if err, ok := m.errors[cmdKey]; ok {
		return "", err
	}

	if output, ok := m.responses[cmdKey]; ok {
		return output, nil
	}

	return "", fmt.Errorf("no mock response configured for command: %s", cmdKey)
}
