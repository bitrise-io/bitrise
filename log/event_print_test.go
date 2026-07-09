package log

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
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
				ExecutionID: "ExecutionId is not needed",
				Position:    0,
				Title:       "xcode-test@4.1.2",
				ID:          "xcode-test",
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
				ExecutionID: "random-uuid",
				Position:    1,
				Title:       "Very long step name - Very long step name - Very long step name - Very long step name - Very long step name",
				ID:          "this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step-this-is-the-step",
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
				ExecutionID: "another-random-uuid",
				Position:    2,
				Title:       "git::https://github.com/org/repo",
				ID:          "https://github.com/org/repo",
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
				Status:       models.StepRunStatusCodeSuccess.String(),
				StatusReason: "",
				Title:        "Normal step name",
				RunTime:      1234567,
				LastStep:     true,
			},
			expectedOutput: []string{
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[32;1m✓\x1b[0m | \x1b[32;1mNormal step name                                             \x1b[0m | 20.6 min |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Failed",
			params: StepFinishedParams{
				Status:   models.StepRunStatusCodeFailed.String(),
				Errors:   []models.StepError{{Code: 1, Message: "exit code: 1"}},
				Title:    "Loooooooooooooooooooooooooooooooooong step name",
				RunTime:  9999,
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[31;1mexit code: 1\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1mx\x1b[0m | \x1b[31;1mLoooooooooooooooooooooooooooooooooong step name (Failed)     \x1b[0m | 10.00 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Failed skippable",
			params: StepFinishedParams{
				Status:       models.StepRunStatusCodeFailedSkippable.String(),
				StatusReason: `This Step failed, but it was marked as "is_skippable"", so the build continued.`,
				Errors:       []models.StepError{{Code: 1, Message: "exit code: 3"}},
				Title:        "Simple Git",
				RunTime:      3333,
				LastStep:     true,
			},
			expectedOutput: []string{
				"\x1b[31;1mexit code: 3\x1b[0m",
				"\x1b[34;1mThis Step failed, but it was marked as \"is_skippable\"\", so the build continued.\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[33;1m!\x1b[0m | \x1b[33;1mSimple Git (Failed)                                          \x1b[0m | 3.33 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Skipped",
			params: StepFinishedParams{
				Status:       models.StepRunStatusCodeSkipped.String(),
				StatusReason: `This Step was skipped, because a previous Step failed, and this Step was not marked "is_always_run".`,
				Title:        "Step",
				RunTime:      654321,
				LastStep:     true,
			},
			expectedOutput: []string{
				"\x1b[34;1mThis Step was skipped, because a previous Step failed, and this Step was not marked \"is_always_run\".\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[34;1m-\x1b[0m | \x1b[34;1mStep (Skipped)                                               \x1b[0m | 10.9 min |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Skipped with run if",
			params: StepFinishedParams{
				Status: models.StepRunStatusCodeSkippedWithRunIf.String(),
				StatusReason: `This Step was skipped, because its "run_if" expression evaluated to false.
The "run_if" expression was: false`,
				Title:    "Step",
				RunTime:  42424242,
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[34;1mThis Step was skipped, because its \"run_if\" expression evaluated to false.\nThe \"run_if\" expression was: false\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[34;1m-\x1b[0m | \x1b[34;1mStep (Skipped)                                               \x1b[0m | 12 hour |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Preparation failed",
			params: StepFinishedParams{
				Status:   models.StepRunStatusCodePreparationFailed.String(),
				Errors:   []models.StepError{{Code: 1, Message: "Preparing Step (Step) failed: exit code: 3"}},
				Title:    "Step",
				RunTime:  11111,
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[31;1mPreparing Step (Step) failed: exit code: 3\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1mx\x1b[0m | \x1b[31;1mStep (Failed)                                                \x1b[0m | 11.11 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Aborted with custom timeout",
			params: StepFinishedParams{
				Status:   models.StepRunStatusAbortedWithCustomTimeout.String(),
				Errors:   []models.StepError{{Code: 1, Message: "This Step timed out after 5m."}},
				Title:    "Step",
				RunTime:  99099,
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[31;1mThis Step timed out after 5m.\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1m/\x1b[0m | \x1b[31;1mStep (Failed)                                                \x1b[0m | 1.7 min |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Aborted with no output",
			params: StepFinishedParams{
				Status:   models.StepRunStatusAbortedWithNoOutputTimeout.String(),
				Errors:   []models.StepError{{Code: 1, Message: "This Step failed, because it has not sent any output for 4m."}},
				Title:    "Step",
				RunTime:  101,
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[31;1mThis Step failed, because it has not sent any output for 4m.\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1m/\x1b[0m | \x1b[31;1mStep (Failed)                                                \x1b[0m | 0.10 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Error status prints the issue and source url",
			params: StepFinishedParams{
				Status:        models.StepRunStatusCodeFailed.String(),
				Errors:        []models.StepError{{Code: 11, Message: "This is an error message"}},
				Title:         "Failed step",
				RunTime:       88888,
				SupportURL:    "https://issue-url-issue-url-issue-url-issue-url-issue-url-issue-url-issue-url-issue-url",
				SourceCodeURL: "https://source-code-url",
				LastStep:      true,
			},
			expectedOutput: []string{
				"\x1b[31;1mThis is an error message\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1mx\x1b[0m | \x1b[31;1mFailed step (Failed)                                         \x1b[0m | 1.5 min |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: https://issue-url-issue-url-issue-url-issue-url-issue-url-... |",
				"| Source: https://source-code-url                                              |",
				"+---+---------------------------------------------------------------+----------+",
			},
		},
		{
			name: "Step update info is printed in the footer",
			params: StepFinishedParams{
				Status:       models.StepRunStatusCodeSuccess.String(),
				StatusReason: "",
				Title:        "Step",
				RunTime:      65748,
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
				"| \x1b[32;1m✓\x1b[0m | \x1b[32;1mStep                                                         \x1b[0m | 1.1 min |",
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
				Status:  models.StepRunStatusCodeFailed.String(),
				Errors:  []models.StepError{{Code: 42, Message: "exit code: 42"}},
				Title:   "Loooooooooong step naaaaaaaaaaaaaaaaaaaaaaaaaaaaaaame",
				RunTime: 223,
				Deprecation: &StepDeprecation{
					RemovalDate: "2022-10-26",
					Note:        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. In at ipsum nec orci convallis efficitur. Nulla ultrices eros non nisi tempus feugiat. Donec ac sapien in odio ultrices ullamcorper vel id erat. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed sed placerat augue, tincidunt varius ipsum. Donec.",
				},
				LastStep: true,
			},
			expectedOutput: []string{
				"\x1b[31;1mexit code: 42\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[31;1mx\x1b[0m | \x1b[31;1m[Deprecated]\x1b[0m \x1b[31;1mLoooooooooong step naaaaaaaaaaaaaaaa... (Failed)\x1b[0m | 0.22 sec |",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"|                                                                              |",
				"| \x1b[31;1mRemoval date:\x1b[0m 2022-10-26                                                     |",
				"| \x1b[31;1mRemoval notes:\x1b[0m Lorem ipsum dolor sit amet, consectetur adipiscing elit. In   |",
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
				Status:  models.StepRunStatusCodeSuccess.String(),
				Title:   "Regular step",
				RunTime: 111111111111,
				Errors:  []models.StepError{{Code: 11, Message: "This is an error message"}},
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
				"\x1b[31;1mThis is an error message\x1b[0m",
				"|                                                                              |",
				"+---+---------------------------------------------------------------+----------+",
				"| \x1b[32;1m✓\x1b[0m | \x1b[31;1m[Deprecated]\x1b[0m \x1b[32;1mRegular step                                    \x1b[0m | 999+ hour|",
				"+---+---------------------------------------------------------------+----------+",
				"| Issue tracker: \x1b[33;1mNot provided\x1b[0m                                                  |",
				"| Source: \x1b[33;1mNot provided\x1b[0m                                                         |",
				"|                                                                              |",
				"| Update available: 1 (1.2.3) -> 9.9.9                                         |",
				"| Release notes are available below                                            |",
				"| https://releases-url                                                         |",
				"|                                                                              |",
				"| \x1b[31;1mRemoval date:\x1b[0m 2022-10-26                                                     |",
				"| \x1b[31;1mRemoval notes:\x1b[0m This is deprecated                                            |",
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
