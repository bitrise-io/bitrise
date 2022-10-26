package log

import (
	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogMessageWidthIsAboveTheMinimumValue(t *testing.T) {
	// This test is here because of a comment above the constant that it should be at least 45 char wide.
	assert.True(t, stepRunSummaryBoxWidthInChars > 45)
}

func TestStepHeaderPrinting(t *testing.T) {
	tests := []struct {
		name           string
		params         StepStartedParams
		expectedOutput []string
	}{
		{
			name: "Only prints the values which need to appear in the console output",
			params: StepStartedParams{
				ExecutionId: "ExecutionId is not needed",
				Position:    0,
				Title:       "xcode-test@4.1.2",
				Id:          "xcode-test",
				Version:     "4.1.2",
				Collection:  "Steplib",
				Toolkit:     "Go",
				StartTime:   "2022-10-19T10:28:33Z ",
			},
			expectedOutput: []string{
				"+------------------------------------------------------------------------------+",
				"| (0) xcode-test@4.1.2                                                         |",
				"+------------------------------------------------------------------------------+",
				"| id: xcode-test                                                               |",
				"| version: 4.1.2                                                               |",
				"| collection: Steplib                                                          |",
				"| toolkit: Go                                                                  |",
				"| time: 2022-10-19T10:28:33Z                                                   |",
				"+------------------------------------------------------------------------------+",
				"|                                                                              |",
			},
		},
		{
			name: "Long step parameter values are truncated",
			params: StepStartedParams{
				ExecutionId: "random-uuid",
				Position:    1,
				Title:       "Very long step name - Very long step name - Very long step name - Very long step name - Very long step name",
				Id:          "this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step",
				Version:     "1.1.2",
				Collection:  "Steplib",
				Toolkit:     "Go",
				StartTime:   "Now",
			},
			expectedOutput: []string{
				"+------------------------------------------------------------------------------+",
				"| (1) Very long step name - Very long step name - Very long step name - Ver... |",
				"+------------------------------------------------------------------------------+",
				"| id: this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-t... |",
				"| version: 1.1.2                                                               |",
				"| collection: Steplib                                                          |",
				"| toolkit: Go                                                                  |",
				"| time: Now                                                                    |",
				"+------------------------------------------------------------------------------+",
				"|                                                                              |",
			},
		},
		{
			name: "Prints empty fields",
			params: StepStartedParams{
				ExecutionId: "another-random-uuid",
				Position:    2,
				Title:       "git::https://github.com/org/repo",
				Id:          "https://github.com/org/repo",
				Version:     "",
				Collection:  "Git",
				Toolkit:     "",
				StartTime:   "42",
			},
			expectedOutput: []string{
				"+------------------------------------------------------------------------------+",
				"| (2) git::https://github.com/org/repo                                         |",
				"+------------------------------------------------------------------------------+",
				"| id: https://github.com/org/repo                                              |",
				"| version:                                                                     |",
				"| collection: Git                                                              |",
				"| toolkit:                                                                     |",
				"| time: 42                                                                     |",
				"+------------------------------------------------------------------------------+",
				"|                                                                              |",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedOutput, generateStepStartedHeaderLines(test.params))
		})
	}
}

func TestStepFooterPrinting(t *testing.T) {
	tests := []struct {
		name           string
		params         StepFinishedParams
		expectedOutput []string
	}{
		{
			name: "Success",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeSuccess,
				Status:         "success",
				Title:          "Normal step name",
				RunTime:        1234567,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[32;1m✓\u001B[0m | \u001B[32;1mNormal step name                                             \u001B[0m | 20.6 min |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Failed",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeFailed,
				Status:         "failed",
				StatusReason:   "exit code: 1",
				Title:          "Loooooooooooooooooooooooooooooooooong step name",
				RunTime:        9999,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1mx\u001B[0m | \u001B[31;1mLoooooooooooooooooooooooooooooooooong step ... (exit code: 1)\u001B[0m | 10.00 sec |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Failed skippable",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeFailedSkippable,
				Status:         "failed_skippable",
				StatusReason:   "exit code: 2",
				Title:          "Simple Git",
				RunTime:        3333,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[33;1m!\u001B[0m | \u001B[33;1mSimple Git (exit code: 2)                                    \u001B[0m | 3.33 sec |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Skipped",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeSkipped,
				Status:         "skipped",
				Title:          "Step",
				RunTime:        654321,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[34;1m-\u001B[0m | \u001B[34;1mStep                                                         \u001B[0m | 10.9 min |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Skipped with run if",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeSkippedWithRunIf,
				Status:         "skipped_with_run_if",
				Title:          "Step",
				RunTime:        42424242,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[34;1m-\u001B[0m | \u001B[34;1mStep                                                         \u001B[0m | 12 hour |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Preparation failed",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodePreparationFailed,
				Status:         "preparation_failed",
				StatusReason:   "exit code: 3",
				Title:          "Step",
				RunTime:        11111,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1mx\u001B[0m | \u001B[31;1mStep (exit code: 3)                                          \u001B[0m | 11.11 sec |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Aborted with custom timeout",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusAbortedWithCustomTimeout,
				Status:         "aborted_with_custom_timeout",
				StatusReason:   "timed out",
				Title:          "Step",
				RunTime:        99099,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1m/\u001B[0m | \u001B[31;1mStep (timed out)                                             \u001B[0m | 1.7 min |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Aborted with no output",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusAbortedWithNoOutputTimeout,
				Status:         "aborted_with_no_output",
				StatusReason:   "timed out due to no output",
				Title:          "Step",
				RunTime:        101,
				LastStep:       true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1m/\u001B[0m | \u001B[31;1mStep (timed out due to no output)                            \u001B[0m | 0.10 sec |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Error status prints the issue and source url",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeFailed,
				Status:         "failed",
				StatusReason:   "exit code: 11",
				Title:          "Failed step",
				RunTime:        88888,
				SupportURL:     "https://issue-url-issue-url-issue-url-issue-url-issue-url-issue-url-issue-url-issue-url",
				SourceCodeURL:  "https://source-code-url",
				Errors: []StepError{
					{Code: 11, Message: "This is an error message"},
				},
				LastStep: true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1mx\u001B[0m | \u001B[31;1mFailed step (exit code: 11)                                  \u001B[0m | 1.5 min |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: https://issue-url-issue-url-issue-url-issue-url-issue-url-... |",
				"| Source: https://source-code-url                                              |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Step update info is printed in the footer",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeSuccess,
				Status:         "success",
				Title:          "Step",
				RunTime:        65748,
				Update: &StepUpdate{
					OriginalVersion: "1",
					ResolvedVersion: "1.2.3",
					LatestVersion:   "9.9.9",
					ReleasesURL:     "https://releases-url",
				},
				LastStep: true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[32;1m✓\u001B[0m | \u001B[32;1mStep                                                         \u001B[0m | 1.1 min |",
				"+---+---------------------------------------------------------------+----------+",
				"| Update available: 1 (1.2.3) -> 9.9.9                                         |",
				"| Release notes are available below                                            |",
				"| https://releases-url                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Deprecation is printed in the footer",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeFailed,
				Status:         "failed",
				StatusReason:   "exit code: 42",
				Title:          "Loooooooooong step naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaame",
				RunTime:        223,
				Deprecation: &StepDeprecation{
					RemovalDate: "2022-10-26",
					Note:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. In at ipsum nec orci convallis efficitur. Nulla ultrices eros non nisi tempus feugiat. Donec ac sapien in odio ultrices ullamcorper vel id erat. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed sed placerat augue, tincidunt varius ipsum. Donec.",
				},
				LastStep: true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1mx\u001B[0m | \u001B[31;1m[Deprecated]\u001B[0m \u001B[31;1mLoooooooooong step naaaaaaaaa... (exit code: 42)\u001B[0m | 0.22 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[31;1mRemoval date:\u001B[0m 2022-10-26                                                     |",
				"| \u001B[31;1mRemoval notes:\u001B[0m Lorem ipsum dolor sit amet, consectetur adipiscing elit. In   |",
				"| at ipsum nec orci convallis efficitur. Nulla ultrices eros non nisi tempus   |",
				"| feugiat. Donec ac sapien in odio ultrices ullamcorper vel id erat. Interdum  |",
				"| et malesuada fames ac ante ipsum primis in faucibus. Sed sed placerat        |",
				"| augue, tincidunt varius ipsum. Donec.                                        |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Error urls, update and deprecation info can appear at the same time",
			params: StepFinishedParams{
				InternalStatus: models.StepRunStatusCodeSuccess,
				Status:         "success",
				Title:          "Regular step",
				RunTime:        111111111111,
				Errors: []StepError{
					{Code: 11, Message: "This is an error message"},
				},
				Update: &StepUpdate{
					OriginalVersion: "1",
					ResolvedVersion: "1.2.3",
					LatestVersion:   "9.9.9",
					ReleasesURL:     "https://releases-url",
				},
				Deprecation: &StepDeprecation{
					RemovalDate: "2022-10-26",
					Note:        "This is deprecated",
				},
				LastStep: true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \u001B[32;1m✓\u001B[0m | \u001B[31;1m[Deprecated]\u001B[0m \u001B[32;1mRegular step                                    \u001B[0m | 999+ hour|",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \u001B[33;1mNot provided\u001B[0m                                                  |",
				"| Source: \u001B[33;1mNot provided\u001B[0m                                                         |",
				"|                                                                              |",
				"| Update available: 1 (1.2.3) -> 9.9.9                                         |",
				"| Release notes are available below                                            |",
				"| https://releases-url                                                         |",
				"|                                                                              |",
				"| \u001B[31;1mRemoval date:\u001B[0m 2022-10-26                                                     |",
				"| \u001B[31;1mRemoval notes:\u001B[0m This is deprecated                                            |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lines := generateStepFinishedFooterLines(test.params)
			assert.Equal(t, test.expectedOutput, lines)
		})
	}
}
