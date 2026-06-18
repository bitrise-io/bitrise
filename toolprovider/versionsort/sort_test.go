package versionsort

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortSemverDescending(t *testing.T) {
	t.Run("sorts semver versions newest first", func(t *testing.T) {
		input := []string{"1.0.0", "3.0.0", "2.0.0"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"3.0.0", "2.0.0", "1.0.0"}, result)
	})

	t.Run("handles pre-release versions", func(t *testing.T) {
		input := []string{"1.0.0", "2.0.0-rc.1", "2.0.0", "1.0.0-beta.1"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"2.0.0", "2.0.0-rc.1", "1.0.0", "1.0.0-beta.1"}, result)
	})

	t.Run("places non-semver after semver", func(t *testing.T) {
		input := []string{"nightly", "2.0.0", "1.0.0", "latest", "3.15.0a8"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"3.15.0a8", "2.0.0", "1.0.0", "nightly", "latest"}, result)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := SortSemverDescending(nil)
		assert.Empty(t, result)
	})

	t.Run("handles single element", func(t *testing.T) {
		input := []string{"1.0.0"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"1.0.0"}, result)
	})

	t.Run("handles all non-semver versions", func(t *testing.T) {
		input := []string{"nightly", "latest", "dev"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"nightly", "latest", "dev"}, result)
	})

	t.Run("handles patch level sorting", func(t *testing.T) {
		input := []string{"1.2.3", "1.2.1", "1.2.2", "1.3.0", "1.1.9"}
		result := SortSemverDescending(input)
		assert.Equal(t, []string{"1.3.0", "1.2.3", "1.2.2", "1.2.1", "1.1.9"}, result)
	})
}
