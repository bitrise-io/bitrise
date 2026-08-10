package baseurl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_AcceptsHTTPS(t *testing.T) {
	u, err := Validate("base URL", "https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", u.String())
}

func TestValidate_RejectsRelative(t *testing.T) {
	_, err := Validate("base URL", "not-a-url")
	assert.ErrorContains(t, err, `base URL "not-a-url" must be an absolute URL`)
}

func TestValidate_RejectsUnparsable(t *testing.T) {
	_, err := Validate("base URL", "://bad")
	assert.Error(t, err)
}

func TestValidate_RejectsPlainHTTP(t *testing.T) {
	_, err := Validate("web_base_url", "http://app.example.com")
	assert.ErrorContains(t, err, `web_base_url "http://app.example.com" must use https (got "http")`)
}

func TestValidate_AllowsPlainHTTPForLoopback(t *testing.T) {
	for _, raw := range []string{"http://localhost:1234", "http://127.0.0.1:1234", "http://[::1]:1234"} {
		_, err := Validate("base URL", raw)
		require.NoError(t, err, raw)
	}
}
