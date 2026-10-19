package asdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReleasedVersionsWithPrefix(t *testing.T) {
	versions := []string{"22.1.0", "22.10.0", "3.4.1", "22.0.0", "jruby-9.4.9.0"}

	assert.Equal(t, []string{"22.10.0", "22.1.0", "22.0.0", "3.4.1", "jruby-9.4.9.0"}, releasedVersionsWithPrefix(versions, ""))
	assert.Equal(t, []string{"22.10.0", "22.1.0"}, releasedVersionsWithPrefix(versions, "22.1"), "a plain string prefix, as asdf resolves it")
	assert.Equal(t, []string{"22.10.0", "22.1.0", "22.0.0"}, releasedVersionsWithPrefix(append(versions, "220.0.0"), "22."), "the prefix is used as given")
	assert.Nil(t, releasedVersionsWithPrefix(versions, "23"))
}
