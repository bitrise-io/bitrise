package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetSetUnset_RoundTrip(t *testing.T) {
	var c Config

	require.NoError(t, c.Set(KeyAPIBaseURL, "https://api.example.com"))
	v, err := c.Get(KeyAPIBaseURL)
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", v)

	require.NoError(t, c.Unset(KeyAPIBaseURL))
	v, err = c.Get(KeyAPIBaseURL)
	require.NoError(t, err)
	assert.Equal(t, "", v)
}

func TestConfig_Get_UnknownKey(t *testing.T) {
	var c Config
	_, err := c.Get("nope")
	assert.ErrorContains(t, err, `unknown config key "nope"`)
}

func TestConfig_Set_UnknownKey(t *testing.T) {
	var c Config
	err := c.Set("nope", "value")
	assert.ErrorContains(t, err, `unknown config key "nope"`)
}

func TestConfig_Set_InvalidURLLeavesConfigUntouched(t *testing.T) {
	c := Config{APIBaseURL: "https://original.example"}

	err := c.Set(KeyAPIBaseURL, "not-a-url")
	assert.ErrorContains(t, err, `api_base_url "not-a-url" must be an absolute URL`)
	assert.Equal(t, "https://original.example", c.APIBaseURL)
}

func TestConfig_Set_WebBaseURL(t *testing.T) {
	var c Config
	require.NoError(t, c.Set(KeyWebBaseURL, "https://app.example.com"))
	assert.Equal(t, "https://app.example.com", c.WebBaseURL)

	err := c.Set(KeyWebBaseURL, "://bad")
	assert.Error(t, err)
	assert.Equal(t, "https://app.example.com", c.WebBaseURL, "invalid Set must not mutate c")
}

func TestConfig_Set_RejectsPlainHTTP(t *testing.T) {
	for _, key := range Keys {
		var c Config
		err := c.Set(key, "http://api.example.com")
		require.Error(t, err, key)
		assert.ErrorContains(t, err, `must use https (got "http")`)
	}
}

func TestConfig_Set_AllowsPlainHTTPForLoopback(t *testing.T) {
	for _, host := range []string{"http://localhost:1234", "http://127.0.0.1:1234", "http://[::1]:1234"} {
		var c Config
		require.NoError(t, c.Set(KeyAPIBaseURL, host), host)
	}
}
