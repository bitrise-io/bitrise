package stepman

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const (
	givenSteplibURI  = "https://github.com/bitrise-io/steplib"
	givenFolderAlias = "12334556"
	givenHomePath    = "/usr/testeruser/"
	givenStepID      = "test-custom-step"
	givenStepVersion = "0.5.6"
)

func GivenRoute() SteplibRoute {
	return SteplibRoute{
		SteplibURI:  givenSteplibURI,
		FolderAlias: givenFolderAlias,
	}
}

func Test_GivenHomeDir_WhenGetStepmanDirPathCalled_ThenGoodPathReturned(t *testing.T) {
	// Given
	oldhome := os.Getenv("HOME")
	err := os.Setenv("HOME", givenHomePath)
	defer func() {
		err := os.Setenv("HOME", oldhome)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	expected := filepath.Join(givenHomePath, ".stepman")

	// When
	actual := GetStepmanDirPath()

	// Then
	assert.Equal(t, expected, actual)
}

func Test_GivenStepmanDir_WhenGetCollectionDirPathCalled_ThenGoodPathReturned(t *testing.T) {
	// Given
	oldhome := os.Getenv("HOME")
	err := os.Setenv("HOME", givenHomePath)
	defer func() {
		err := os.Setenv("HOME", oldhome)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	expected := filepath.Join(GetStepmanDirPath(), "step_collections")

	// When
	actual := GetCollectionsDirPath()

	// Then
	assert.Equal(t, expected, actual)
}

func Test_GivenRoute_WhenGetLibraryBaseDirPathCalled_ThenGoodPathReturned(t *testing.T) {
	// Given
	route := GivenRoute()
	expected := filepath.Join(GetCollectionsDirPath(), route.FolderAlias, "collection")

	// When
	actual := GetLibraryBaseDirPath(route)

	// Then
	assert.Equal(t, expected, actual)
}

func Test_GivenRouteAndStepId_WhenGetStepCollectionDirPath_ThenGoodPathReturned(t *testing.T) {
	// Given
	route := GivenRoute()
	step := givenStepID
	version := givenStepVersion
	expected := filepath.Join(GetLibraryBaseDirPath(route), "steps", step, version)

	// When
	actual := GetStepCollectionDirPath(route, step, version)

	// Then
	assert.Equal(t, expected, actual)
}

func Test_GivenRouteAndStepId_WhenGetStepGlobalInfoPathCalled_ThenGoodPathReturned(t *testing.T) {
	// Given
	route := GivenRoute()
	step := givenStepID
	expected := filepath.Join(GetLibraryBaseDirPath(route), "steps", step, "step-info.yml")

	// When
	actual := GetStepGlobalInfoPath(route, step)

	// Then
	assert.Equal(t, expected, actual)
}
