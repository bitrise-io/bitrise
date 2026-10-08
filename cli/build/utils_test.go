package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildWebURL_EscapesIDs(t *testing.T) {
	assert.Equal(t, "https://app.bitrise.io/app/my-app/build/my-build",
		buildWebURL("https://app.bitrise.io", "my-app", "my-build"))
	assert.Equal(t, "https://app.bitrise.io/app/a%2Fb/build/c%3Fd%23e",
		buildWebURL("https://app.bitrise.io", "a/b", "c?d#e"))
}
