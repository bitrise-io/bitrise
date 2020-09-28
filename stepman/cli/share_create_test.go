package cli

import (
	"github.com/bitrise-io/stepman/models"
	"testing"

	"github.com/stretchr/testify/assert"
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
