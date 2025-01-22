package integration

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func binPath() string {
	return os.Getenv("INTEGRATION_TEST_BINARY_PATH")
}

func toBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func toJSON(t *testing.T, stringStringMap map[string]interface{}) string {
	bytes, err := json.Marshal(stringStringMap)
	require.NoError(t, err)
	return string(bytes)
}

func withRunningTimeCheck(f func()) time.Duration {
	start := time.Now()
	f()
	end := time.Now()

	return end.Sub(start)
}
