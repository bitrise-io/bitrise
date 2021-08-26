package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_SensitiveInputs(t *testing.T) {
	configPth := "sensitive_inputs_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "test-sensitive-env-and-output", "--config", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Equal(t, out, testOutput)
}

const testOutput = `
██████╗ ██╗████████╗██████╗ ██╗███████╗███████╗
██╔══██╗██║╚══██╔══╝██╔══██╗██║██╔════╝██╔════╝
██████╔╝██║   ██║   ██████╔╝██║███████╗█████╗
██╔══██╗██║   ██║   ██╔══██╗██║╚════██║██╔══╝
██████╔╝██║   ██║   ██║  ██║██║███████║███████╗
╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝╚══════╝

[32;1m  version: 1.47.2[0m

[36mINFO[0m[17:06:54] [33;1mbitrise runs in Secret Filtering mode[0m 
[36mINFO[0m[17:06:54] [33;1mbitrise runs in Secret Envs Filtering mode[0m 
[36mINFO[0m[17:06:54] Running workflow: [32;1mtest-sensitive-env-and-output[0m 

[34;1mSwitching to workflow:[0m test-sensitive-env-and-output

[36mINFO[0m[17:06:54] Step uses latest version -- Updating StepLib ... 
+------------------------------------------------------------------------------+
| (0) Add a sensitive env                                                      |
+------------------------------------------------------------------------------+
| id: script                                                                   |
| version: 1.1.6                                                               |
| collection: https://github.com/bitrise-io/bitrise-steplib.git                |
| toolkit: bash                                                                |
| time: 2021-08-26T17:06:57+02:00                                              |
+------------------------------------------------------------------------------+
|                                                                              |
+ bitrise envman add --key MYTESTKEY --value mysupersecret --sensitive
|                                                                              |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mAdd a sensitive env[0m                                           | 2.85 sec |
+---+---------------------------------------------------------------+----------+

										▼

+------------------------------------------------------------------------------+
| (1) Add a step with sensitive output                                         |
+------------------------------------------------------------------------------+
| id: ./test_step_with_sensitive_output                                        |
| version:                                                                     |
| collection: path                                                             |
| toolkit: bash                                                                |
| time: 2021-08-26T17:06:57+02:00                                              |
+------------------------------------------------------------------------------+
|                                                                              |
+ bitrise envman add --key TESTOUTPUT --value myotherverysecret
|                                                                              |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mAdd a step with sensitive output[0m                              | 0.37 sec |
+---+---------------------------------------------------------------+----------+

										▼

+------------------------------------------------------------------------------+
| (2) Try to print sensitive env and sensitive output                          |
+------------------------------------------------------------------------------+
| id: script                                                                   |
| version: 1.1.6                                                               |
| collection: https://github.com/bitrise-io/bitrise-steplib.git                |
| toolkit: bash                                                                |
| time: 2021-08-26T17:06:58+02:00                                              |
+------------------------------------------------------------------------------+
|                                                                              |
[REDACTED]
[REDACTED]
|                                                                              |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mTry to print sensitive env and sensitive output[0m               | 0.69 sec |
+---+---------------------------------------------------------------+----------+


+------------------------------------------------------------------------------+
|                               bitrise summary                                |
+---+---------------------------------------------------------------+----------+
|   | title                                                         | time (s) |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mAdd a sensitive env[0m                                           | 2.85 sec |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mAdd a step with sensitive output[0m                              | 0.37 sec |
+---+---------------------------------------------------------------+----------+
| [32;1m✓[0m | [32;1mTry to print sensitive env and sensitive output[0m               | 0.69 sec |
+---+---------------------------------------------------------------+----------+
| Total runtime: 3.91 sec                                                      |
+------------------------------------------------------------------------------+

[34;1m[0m
[34;1mSubmitting anonymized usage information...[0m
[34;1mFor more information visit:[0m
[34;1mhttps://github.com/bitrise-io/bitrise-plugins-analytics/blob/master/README.md[0m

[32;1mBitrise build successful[0m
`
