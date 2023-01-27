package stepoutput

import (
	"bytes"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/log"

	"github.com/stretchr/testify/require"
)

func Test_writer_Write(t *testing.T) {
	tests := []struct {
		name       string
		messages   []string
		loggerOpts log.LoggerOpts
		secrets    []string
		want       string
	}{
		{
			name:     "Simple log",
			messages: []string{"failed to create file artifact: /bitrise/src/assets"},
			want:     "failed to create file artifact: /bitrise/src/assets",
		},
		{
			name: "error log",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			want: `[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`,
		},
		{
			name: "error log - json",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			loggerOpts: log.LoggerOpts{LoggerType: log.JSONLogger},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets"}
`,
		},
		{
			name: "error log - json - secret filtering",
			messages: []string{`[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m`},
			loggerOpts: log.LoggerOpts{LoggerType: log.JSONLogger},
			secrets:    []string{"SECRET_KEY"},
			want: `{"timestamp":"0001-01-01T00:00:00Z","type":"log","producer":"","level":"error","message":"failed to create file artifact: /bitrise/src/assets:\n  failed to get file size, error: file not exist at: /bitrise/src/assets"}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buff bytes.Buffer
			tt.loggerOpts.Writer = &buff
			tt.loggerOpts.TimeProvider = func() time.Time {
				return time.Time{}
			}
			w := NewWriter(tt.secrets, tt.loggerOpts)
			for _, message := range tt.messages {
				gotN, err := w.Write([]byte(message))
				require.NoError(t, err)
				require.Equal(t, len(message), gotN)
			}
			require.Equal(t, tt.want, string(buff.Bytes()))
		})
	}
}

var failingDeployStepErrorMessages = []string{`failed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets`, `deploy failed, error:
  failed to create file artifact: /bitrise/src/assets:
    failed to get file size, error: file not exist at: /bitrise/src/assets`}

const failingDeployStepLog = `[34;1mCollecting files to deploy...
[0mBuild Artifact deployment mode: deploying single file
List of files to deploy:
- /bitrise/src/assets

[34;1mDeploying files...
[0mDeploying file: /bitrise/src/assets
[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m

[31;1mdeploy failed, error:
  failed to create file artifact: /bitrise/src/assets:
    failed to get file size, error: file not exist at: /bitrise/src/assets[0m`
