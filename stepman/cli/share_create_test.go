package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTag(t *testing.T) {
	cases := []struct {
		tag   string
		valid bool
	}{
		{tag: "1.0.0", valid: true},
		{tag: "1.0", valid: false},
		{tag: "v1.0.0", valid: false},
	}

	for _, tc := range cases {
		got := validateTag(tc.tag)
		valid := got == nil
		assert.Equal(t, tc.valid, valid, "validateTag(%s) == nil should be %t but got %s", tc.tag, tc.valid, got)
	}
}

func TestGetDefaultStepGroupSpec(t *testing.T) {
	// Given
	expected := models.StepGroupInfoModel{
		Maintainer: "community",
	}

	// When
	actual := getDefaultStepGroupSpec()

	// Then
	assert.Equal(t, expected, actual)
}

func TestApplyStepYMLOverride(t *testing.T) {
	cases := []struct {
		name            string
		baseModel       models.StepModel
		overrideContent string
		expectedError   string
		validateResult  func(t *testing.T, result models.StepModel, baseModel models.StepModel)
	}{
		{
			name: "successfully applies override while preserving source and published_at",
			baseModel: models.StepModel{
				Title:   stringPtr("Original Title"),
				Summary: stringPtr("Original summary"),
				Source: &models.StepSourceModel{
					Git:    "https://github.com/original/repo.git",
					Commit: "abc123",
				},
				PublishedAt: timePtr(2024, 1, 1),
			},
			overrideContent: `title: Overridden Title
summary: Overridden summary
description: New description
website: https://github.com/override/repo
`,
			validateResult: func(t *testing.T, result models.StepModel, baseModel models.StepModel) {
				// Verify override fields were applied
				assert.Equal(t, "Overridden Title", *result.Title)
				assert.Equal(t, "Overridden summary", *result.Summary)
				assert.Equal(t, "New description", *result.Description)
				assert.Equal(t, "https://github.com/override/repo", *result.Website)

				// Verify source and published_at were preserved from base model
				require.NotNil(t, result.Source)
				assert.Equal(t, "https://github.com/original/repo.git", result.Source.Git)
				assert.Equal(t, "abc123", result.Source.Commit)
				require.NotNil(t, result.PublishedAt)
				assert.Equal(t, baseModel.PublishedAt, result.PublishedAt)
			},
		},
		{
			name: "preserves nil source and published_at if base has them nil",
			baseModel: models.StepModel{
				Title:       stringPtr("Base Title"),
				Source:      nil,
				PublishedAt: nil,
			},
			overrideContent: `title: Override Title
summary: Override summary
source:
  git: https://github.com/override/source.git
  commit: xyz789
`,
			validateResult: func(t *testing.T, result models.StepModel, baseModel models.StepModel) {
				assert.Equal(t, "Override Title", *result.Title)
				assert.Nil(t, result.Source, "Source should remain nil from base model")
				assert.Nil(t, result.PublishedAt, "PublishedAt should remain nil from base model")
			},
		},
		{
			name: "returns error for invalid YAML",
			baseModel: models.StepModel{
				Title: stringPtr("Base Title"),
			},
			overrideContent: `title: "Unclosed quote
summary: This is invalid YAML
`,
			expectedError: "failed to unmarshal override step.yml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			overridePath := filepath.Join(tmpDir, "override.yml")
			err := os.WriteFile(overridePath, []byte(tc.overrideContent), 0644)
			require.NoError(t, err)

			result, err := applyStepYMLOverride(overridePath, tc.baseModel)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				if tc.validateResult != nil {
					tc.validateResult(t, result, tc.baseModel)
				}
			}
		})
	}
}

// Helper functions for test readability
func stringPtr(s string) *string {
	return &s
}

func timePtr(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}
