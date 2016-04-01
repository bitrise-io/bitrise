package cli

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// should be valid
	bitriseConfContent, warnings, err := generateBitriseYMLContent("App Title", "master")
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))
	require.NotEqual(t, "", bitriseConfContent)
	require.Contains(t, bitriseConfContent, fmt.Sprintf("format_version: %s", models.Version))
	require.Contains(t, bitriseConfContent, `- BITRISE_APP_TITLE: "App Title"`)
	require.Contains(t, bitriseConfContent, `- BITRISE_DEV_BRANCH: "master"`)

	_, warnings, err = bitrise.ConfigModelFromYAMLBytes([]byte(bitriseConfContent))
	require.NoError(t, err)
	require.Equal(t, 0, len(warnings))
}
