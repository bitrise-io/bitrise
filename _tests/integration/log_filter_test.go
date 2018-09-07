package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

const sshKeyLogChunk = `My ssh key: -----BEGIN RSA PRIVATE KEY-----
bitrise_testmfsjOSsRK7+uFr1g4jvIz/yoDrSMRcKrBi/c+iJr+aO8xIb7j2WQ
5F4h84OLEkQEmfsjOSsRK7+uFr1g4jvIz/yoDrSMRcKrBi/c+iJr+aO8xIb7j2WQ
sPXxhoOj4kuoxqFjrQMGyDZ+uIJMD9D+vsov4iDvIBrMkn2TuD/o1X9oISEhDw1l
3tsWqgFxpZprcMw64rdEOJ/7+aJczWvi37kGYjQ4wvSnD+MEoFZIM3fhxDDcxb+I
COjv7Y+Ta++KGjhyu5OJjTAzFyjal0ub0VaVdu8Vg6tAr1grdhQayPYXZqd1TqaU
kniMwxz4hAg+QbhsdSlKzQjgbJJhzn3shiK7kMxL7DrUmhoIgQ1QMUERj4Lt8y9I
J3zHmSq27IEXSzwBIL0JRAsKfcq914f3S2tbyQUi2doJTMxWDgcaL6jkzjCwmCx/
bitrise_testmfsWwlaF+Y0w0xVfAcABHdYjWHx2UHP02EC1ZGUAqF9z6XaCV8l9
oMHHu9lvWKuxpVNPcGY/kR3G897Qn+6vE3yuVwbD4reu0IHAWZzBgt7e3we5
-----END RSA PRIVATE KEY-----`

func Test_LogFilter(t *testing.T) {
	configPth := "log_filter_test_bitrise.yml"
	secretsPth := "log_filter_test_secrets.yml"

	t.Log("trivial test")
	{
		cmd := command.New(binPath(), "run", "primary", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Contains(t, out, `[REDACTED]
[REDACTED]
123454
123453
123452
[REDACTED]`)
		require.NotContains(t, out, `123456
123455
123454
123453
123452
123451`)
	}

	t.Log("multi line test")
	{
		cmd := command.New(binPath(), "run", "test", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Contains(t, out, `My ssh key: [REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]
[REDACTED]`)
		require.NotContains(t, out, sshKeyLogChunk)
	}

	t.Log("newlines in the middle")
	{
		cmd := command.New(binPath(), "run", "newline_test", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Contains(t, out, `SECRET_WITH_NEWLINES_IN_THE_MIDDLE: [REDACTED]
[REDACTED]
[REDACTED]continue the last line`)
		require.NotContains(t, out, sshKeyLogChunk)
	}

	t.Log("newlines at the end")
	{
		cmd := command.New(binPath(), "run", "ending_with_newline_test", "--config", configPth, "--inventory", secretsPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Contains(t, out, `SECRET_ENDING_WITH_NEWLINE: [REDACTED]
starts in a new line`)
		require.NotContains(t, out, sshKeyLogChunk)
	}

	t.Log("disable filtering test")
	{
		secretsPth = "log_filter_disabled_test_secrets.yml"

		cmd := command.New(binPath(), "run", "test", "--config", configPth, "--inventory", secretsPth, "--secret-filtering=false")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Contains(t, out, sshKeyLogChunk)
	}
}
