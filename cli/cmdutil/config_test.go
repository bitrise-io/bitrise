package cmdutil

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestCheckFormatVersionSupported(t *testing.T) {
	t.Run("supported version passes", func(t *testing.T) {
		assert.NoError(t, CheckFormatVersionSupported(models.FormatVersion))
	})

	t.Run("older version passes", func(t *testing.T) {
		assert.NoError(t, CheckFormatVersionSupported("1"))
	})

	t.Run("newer version errors", func(t *testing.T) {
		err := CheckFormatVersionSupported("99")
		assert.ErrorContains(t, err, "higher format version")
	})

	t.Run("unparsable version errors", func(t *testing.T) {
		err := CheckFormatVersionSupported("not-a-version")
		assert.ErrorContains(t, err, "failed to parse bitrise.yml format version")
	})
}
