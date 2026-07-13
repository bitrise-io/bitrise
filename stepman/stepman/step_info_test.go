package stepman

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStepTitle(t *testing.T) {
	infoWithTitle := func(title *string) models.StepInfoModel {
		return models.StepInfoModel{ID: "my-step", Step: models.StepModel{Title: title}}
	}

	t.Run("nil title defaults to the step ID", func(t *testing.T) {
		got := defaultStepTitle(infoWithTitle(nil))
		require.NotNil(t, got.Step.Title)
		assert.Equal(t, "my-step", *got.Step.Title)
	})

	t.Run("empty title defaults to the step ID", func(t *testing.T) {
		got := defaultStepTitle(infoWithTitle(pointers.NewStringPtr("")))
		require.NotNil(t, got.Step.Title)
		assert.Equal(t, "my-step", *got.Step.Title)
	})

	t.Run("existing title is preserved", func(t *testing.T) {
		got := defaultStepTitle(infoWithTitle(pointers.NewStringPtr("Real Title")))
		require.NotNil(t, got.Step.Title)
		assert.Equal(t, "Real Title", *got.Step.Title)
	})
}
