package stepruncmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdSecretRedactionEnabled(t *testing.T) {
	failingBashCmd := `echo -e "\033[31;1mInvalid password: 1234\033[0m"; exit 1`
	secrets := []string{"1234"}
	cmd := New("bash", []string{"-c", failingBashCmd}, "", nil, secrets, 0, 0, nil)
	_, err := cmd.Run()
	require.EqualError(t, err, "Invalid password: [REDACTED]")
}

func TestCmdSecretRedactionDisabled(t *testing.T) {
	failingBashCmd := `echo -e "\033[31;1mInvalid password: 1234\033[0m"; exit 1`
	secrets := []string(nil)
	cmd := New("bash", []string{"-c", failingBashCmd}, "", nil, secrets, 0, 0, nil)
	_, err := cmd.Run()
	require.EqualError(t, err, "Invalid password: 1234")
}
