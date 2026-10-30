package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/internal/update"
)

func TestNewMajorNotice(t *testing.T) {
	notice := newMajorNotice(update.Available{
		Version: "3.0.0",
		URL:     "https://github.com/bitrise-io/bitrise/releases/tag/v3.0.0",
	})

	require.Equal(t, `Bitrise CLI 3.0.0 is available (new major version)
It is not installed automatically, because a major version can contain breaking changes.
Release notes and migration guide: https://github.com/bitrise-io/bitrise/releases/tag/v3.0.0
To install it: bitrise update --version 3.0.0`, notice)
}
