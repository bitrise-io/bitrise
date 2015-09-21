package cli

import (
	"os"
	"path"
	"testing"

	"github.com/bitrise-io/envman/envman"
	"github.com/stretchr/testify/require"
)

func TestPrint(t *testing.T) {
	envsStr := `
envs:
- TEST_HOME1: $HOME
- TEST_HOME2: $TEST_HOME1/test
`
	environments, err := envman.ParseEnvsYML([]byte(envsStr))
	require.Equal(t, nil, err)

	envsJSONList, err := convertToEnsJSONModel(environments, false)
	require.Equal(t, nil, err)
	require.Equal(t, map[string]string{"TEST_HOME1": "$HOME"}, envsJSONList[0])
	require.Equal(t, map[string]string{"TEST_HOME2": "$TEST_HOME1/test"}, envsJSONList[1])

	testHome1 := os.Getenv("HOME")
	testHome2 := path.Join(testHome1, "test")
	envsJSONList, err = convertToEnsJSONModel(environments, true)
	require.Equal(t, nil, err)
	require.Equal(t, map[string]string{"TEST_HOME1": testHome1}, envsJSONList[0])
	require.Equal(t, map[string]string{"TEST_HOME2": testHome2}, envsJSONList[1])
}
