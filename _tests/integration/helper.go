package integration

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
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

func withRunningTimeCheck(f func(), ms time.Duration) time.Duration {
	start := time.Now()
	f()
	end := time.Now()
 
	return end.Sub(start)
}

func assertRunTime(t *testing.T, elapsed time.Duration, limit time.Duration) {
		fastEnough := elapsed < limit
		if (!fastEnough) {
			t.Logf("test case too slow: %s is %s above limit", elapsed, elapsed - limit)
		}
		assert.Equal(t, true, elapsed < limit)
}
