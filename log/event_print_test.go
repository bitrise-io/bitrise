package log

import (
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
