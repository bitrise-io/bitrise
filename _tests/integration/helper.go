package integration

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func binPath() string {
	return os.Getenv("INTEGRATION_TEST_BINARY_PATH")
}

func toBase64(t *testing.T, str string) string {
	bytes := base64.StdEncoding.EncodeToString([]byte(str))
	return string(bytes)
}

func toJSON(t *testing.T, stringStringMap map[string]string) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func copyFile(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}
