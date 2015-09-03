package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// should be valid
	bitriseConfContent, err := generateBitriseYMLContent("App Title", "master")
	require.NoError(t, err)
	require.NotEqual(t, "", bitriseConfContent)

	bitriseConfModel, err := bitrise.ConfigModelFromYAMLBytes([]byte(bitriseConfContent))
	require.NoError(t, err)

	err = bitriseConfModel.Validate()
	require.NoError(t, err)
}
