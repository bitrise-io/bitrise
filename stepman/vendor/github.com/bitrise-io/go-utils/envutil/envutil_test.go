package envutil

import "testing"
import "os"
import "github.com/stretchr/testify/require"
import "github.com/bitrise-io/go-utils/pointers"

func TestSetenvForFunction(t *testing.T) {
	// set an original value
	testKey := "KEY_SetenvForFunction"
	require.NoError(t, os.Setenv(testKey, "orig value"))

	// quick test it
	require.EqualValues(t, "orig value", os.Getenv(testKey))

	// now apply another value, but just for the function
	setEnvErr := SetenvForFunction(testKey, "temp value", func() {
		require.EqualValues(t, "temp value", os.Getenv(testKey))
	})
	require.NoError(t, setEnvErr)

	// should be the original value again
	require.EqualValues(t, "orig value", os.Getenv(testKey))
}

func TestRevokableSetenv(t *testing.T) {
	// set an original value
	testKey := "KEY_RevokableSetenv"
	require.NoError(t, os.Setenv(testKey, "RevokableSetenv orig value"))

	// quick test it
	require.EqualValues(t, "RevokableSetenv orig value", os.Getenv(testKey))

	// revokable set
	revokeFn, err := RevokableSetenv(testKey, "revokable value")
	require.NoError(t, err)

	// env should now be the changed value
	require.EqualValues(t, "revokable value", os.Getenv(testKey))

	// revoke it
	require.NoError(t, revokeFn())

	// should be the original value again
	require.EqualValues(t, "RevokableSetenv orig value", os.Getenv(testKey))
}

func TestStringFlagOrEnv(t *testing.T) {
	testEnvKey := "KEY_TestStringFlagOrEnv"

	revokeFn, err := RevokableSetenv(testEnvKey, "env value")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, revokeFn())
	}()

	// quick test it
	require.EqualValues(t, "env value", os.Getenv(testEnvKey))

	// flag provided - value should be that
	require.Equal(t, "flag value", StringFlagOrEnv(pointers.NewStringPtr("flag value"), testEnvKey))

	// flag not provided - value should be the env's value
	require.Equal(t, "env value", StringFlagOrEnv(nil, testEnvKey))

	// flag provided but empty string - value should be the env's value, it's the same as a nil flag
	require.Equal(t, "env value", StringFlagOrEnv(pointers.NewStringPtr(""), testEnvKey))
}

func TestGetenvWithDefault(t *testing.T) {
	testEnvKey := "KEY_TestGetenvWithDefault"

	// no env set yet, return with default
	require.Equal(t, "default value", GetenvWithDefault(testEnvKey, "default value"))

	// set the env
	revokeFn, err := RevokableSetenv(testEnvKey, "env value")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, revokeFn())
	}()

	// env set - value should be the env's value
	require.Equal(t, "env value", GetenvWithDefault(testEnvKey, "default value"))
}
