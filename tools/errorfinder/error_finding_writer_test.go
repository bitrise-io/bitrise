package errorfinder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_errorFindingWriter_findString(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		want   []ErrorMessage
	}{
		{
			name: "No color string",
			inputs: []string{
				"Test input",
				"newline\nfoo",
			},
			want: nil,
		},
		{
			name: "Black color string",
			inputs: []string{
				"\x1b[30;1mTest input",
				"newline\nfoo\x1b[0m",
			},
			want: nil,
		},
		{
			name: "Simple red string without modifier",
			inputs: []string{
				"\x1b[31mTest input\x1b[0m",
			},
			want: []ErrorMessage{{Message: "Test input"}},
		},
		{
			name: "Simple red string",
			inputs: []string{
				"\x1b[31;1mTest input\x1b[0m",
			},
			want: []ErrorMessage{{Message: "Test input"}},
		},
		{
			name: "Empty red string",
			inputs: []string{
				"Foo\x1b[31;1m\x1b[0mBar",
			},
			want: nil,
		},
		{
			name: "Postfix red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0m",
			},
			want: []ErrorMessage{{Message: "Bar"}},
		},
		{
			name: "Prefix red string",
			inputs: []string{
				"\x1b[31;1mFoo\x1b[0mBar",
			},
			want: []ErrorMessage{{Message: "Foo"}},
		},
		{
			name: "Surrounded red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0mBaz",
			},
			want: []ErrorMessage{{Message: "Bar"}},
		},
		{
			name: "Multiline red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\nBaz\nQux\x1b[0mTest",
			},
			want: []ErrorMessage{{Message: "Bar\nBaz\nQux"}},
		},
		{
			name: "Split red string at content",
			inputs: []string{
				"Foo\x1b[31;1mBa", "r\nBaz\nQux\x1b[0mTest",
			},
			want: []ErrorMessage{{Message: "Bar\nBaz\nQux"}},
		},
		{
			name: "Split red string at control",
			inputs: []string{
				"Foo\x1b", "[31", ";1mBar\nBaz\nQux\x1b[0mTest",
			},
			want: []ErrorMessage{{Message: "Bar\nBaz\nQux"}},
		},
		{
			name: "Red then black",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[30;1mBaz\x1b[0mQux",
			},
			want: []ErrorMessage{{Message: "Bar"}},
		},
		{
			name: "Multiple red sections",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0mBaz\x1b[31;1mQux\x1b[0m",
			},
			want: []ErrorMessage{{Message: "Bar"},
				{Message: "Qux"}},
		},
		{
			name: "Complex multiple red sections",
			inputs: []string{
				"Foo\x1b[", "31;1mB\na\nr\x1b", "[0mBaz\x1b[31;1mQ", "\nu\nx\x1b[0mTest",
			},
			want: []ErrorMessage{{Message: "B\na\nr"},
				{Message: "Q\nu\nx"}},
		},
		{
			name: "Endless red",
			inputs: []string{
				"\x1b[31;1mTest\n in", "put",
			},
			want: []ErrorMessage{{Message: "Test\n input"}},
		},
		{
			name: "Repeated reds",
			inputs: []string{
				"\x1b[31;1mTest \x1b[31;1min", "put\x1b[0m",
			},
			want: []ErrorMessage{{Message: "Test input"}},
		},
		{
			name: "Endless repeated reds",
			inputs: []string{
				"Foo\n\n\n\x1b[31;1mTest \x1b[31;1mi\x1b[31;1mn", "put",
			},
			want: []ErrorMessage{{Message: "Test input"}},
		},
		{
			name: "Multiple control expression",
			inputs: []string{
				"Foo\n\n\n\x1b[1;3;4;31mTest \x1b[31;1mi\x1b[31;1mn", "put",
			},
			want: []ErrorMessage{{Message: "Test input"}},
		},
		{
			name: "Failing deploy step",
			inputs: []string{
				failingDeployStepLog,
			},
			want: []ErrorMessage{{Message: failingDeployStepErrorMessages[0]},
				{Message: failingDeployStepErrorMessages[1]}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewErrorFinder(nil, func() time.Time {
				// UnixNano() is 0 for this time
				return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			})
			for _, input := range tt.inputs {
				_, err := w.Write([]byte(input))
				require.NoError(t, err)
			}
			got := w.ErrorMessages()
			require.Equal(t, tt.want, got)
			//if (tt.want == nil && got != nil) || (tt.want != nil && got == nil) || (tt.want != nil && tt.want.Message != got.Message) {
			//	t.Errorf("got %v. want %v", got, tt.want)
			//}
		})
	}
}

var failingDeployStepErrorMessages = []string{`failed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets`, `deploy failed, error:
  failed to create file artifact: /bitrise/src/assets:
    failed to get file size, error: file not exist at: /bitrise/src/assets`}

const failingDeployStepLog = `[34;1mCollecting files to deploy...
[0mBuild Artifact deployment mode: deploying single file
List of files to deploy:
- /bitrise/src/assets

[34;1mDeploying files...
[0mDeploying file: /bitrise/src/assets
[31;1mfailed to create file artifact: /bitrise/src/assets:
  failed to get file size, error: file not exist at: /bitrise/src/assets[0m

[31;1mdeploy failed, error:
  failed to create file artifact: /bitrise/src/assets:
    failed to get file size, error: file not exist at: /bitrise/src/assets[0m`
