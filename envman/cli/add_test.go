package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestReadMax(t *testing.T) {
	{
		s, err := readMax(strings.NewReader(""), 20*1024)
		require.NoError(t, err)
		require.Equal(t, "", s, fmt.Sprintf("s was: (%s)", s))
	}

	{
		s, err := readMax(strings.NewReader("test"), 20*1024)
		require.NoError(t, err)
		require.Equal(t, "test", s, fmt.Sprintf("s was: (%s)", s))
	}

	{
		s, err := readMax(strings.NewReader("test"), 2)
		require.NoError(t, err)
		require.Equal(t, "te", s, fmt.Sprintf("s was: (%s)", s))
	}
}

func TestReadMaxWithTimeout(t *testing.T) {
	t.Log("read empty string")
	{
		s, err := readMaxWithTimeout(strings.NewReader(""), 20*1024, 1*time.Second)
		require.NoError(t, err)
		require.Equal(t, "", s, fmt.Sprintf("s was: (%s)", s))
	}

	t.Log("read some string")
	{
		s, err := readMaxWithTimeout(strings.NewReader("some text to be read"), 20*1024, 1*time.Second)
		require.NoError(t, err)
		require.Equal(t, "some text to be read", s, fmt.Sprintf("s was: (%s)", s))
	}

	t.Log("reading from closed pipe, with data")
	{
		r, w := io.Pipe()

		// writing without a reader will deadlock so write in a goroutine
		go func() {
			defer func() { require.NoError(t, w.Close()) }()
			_, err := fmt.Fprint(w, "some text to be read")
			require.NoError(t, err)
		}()

		s, err := readMaxWithTimeout(r, 20*1024, 1*time.Second)
		require.NoError(t, err)
		require.Equal(t, "some text to be read", s, fmt.Sprintf("s was: (%s)", s))
	}

	t.Log("reading from closed pipe, with delayed data - timeout")
	{
		r, w := io.Pipe()

		go func() {
			time.Sleep(2 * time.Second)
			defer func() { require.NoError(t, w.Close()) }()
			_, err := fmt.Fprint(w, "some text to be read")
			require.NoError(t, err)
		}()

		s, err := readMaxWithTimeout(r, 20*1024, 1*time.Second)
		require.EqualError(t, err, errTimeout.Error())
		require.Equal(t, "", s, fmt.Sprintf("s was: (%s)", s))
	}

	t.Log("reading from unclosed pipe, with data")
	{
		r, w := io.Pipe()

		go func() {
			_, err := fmt.Fprint(w, "some text to be read")
			require.NoError(t, err)
		}()

		s, err := readMaxWithTimeout(r, 20*1024, 1*time.Second)
		require.NoError(t, err)
		require.Equal(t, "some text to be read", s, fmt.Sprintf("s was: (%s)", s))
	}

	t.Log("reading from unclosed pipe, without data - timeout")
	{
		r, _ := io.Pipe()

		s, err := readMaxWithTimeout(r, 20*1024, 1*time.Second)
		require.EqualError(t, err, errTimeout.Error())
		require.Equal(t, "", s, fmt.Sprintf("s was: (%s)", s))
	}
}

func TestEnvListSizeInBytes(t *testing.T) {
	str100Bytes := strings.Repeat("a", 100)
	require.Equal(t, 100, len([]byte(str100Bytes)))

	env := models.EnvironmentItemModel{
		"key": str100Bytes,
	}

	envList := []models.EnvironmentItemModel{env}
	size, err := envListSizeInBytes(envList)
	require.Equal(t, nil, err)
	require.Equal(t, 100, size)

	envList = []models.EnvironmentItemModel{env, env}
	size, err = envListSizeInBytes(envList)
	require.Equal(t, nil, err)
	require.Equal(t, 200, size)
}

func TestValidateEnv(t *testing.T) {
	// Valid - max allowed
	str20KBytes := strings.Repeat("a", (20 * 1024))
	env1 := models.EnvironmentItemModel{
		"key": str20KBytes,
	}
	envs := []models.EnvironmentItemModel{env1}

	valValue, err := validateEnv("key", str20KBytes, envs)
	require.NoError(t, err)
	require.Equal(t, str20KBytes, valValue)

	// List oversize
	//  first create a large, but valid env set
	for i := 0; i < 3; i++ {
		envs = append(envs, env1)
	}

	valValue, err = validateEnv("key", str20KBytes, envs)
	require.NoError(t, err)
	require.Equal(t, str20KBytes, valValue)

	// append one more -> too large
	envs = append(envs, env1)
	_, err = validateEnv("key", str20KBytes, envs)
	require.Equal(t, errors.New("environment list too large"), err)

	// List oversize + too big value
	str10Kbytes := strings.Repeat("a", (10 * 1024))
	env1 = models.EnvironmentItemModel{
		"key": str10Kbytes,
	}
	envs = []models.EnvironmentItemModel{}
	for i := 0; i < 8; i++ {
		env := models.EnvironmentItemModel{
			"key": str10Kbytes,
		}
		envs = append(envs, env)
	}

	str21Kbytes := strings.Repeat("a", (21 * 1024))

	valValue, err = validateEnv("key", str21Kbytes, envs)
	require.NoError(t, err)
	require.Equal(t, "environment value too large - rejected", valValue)
}
