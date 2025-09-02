package mise

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestFallbackDownloadURLs(t *testing.T) {
	tests := []struct {
		os   string
		arch string
	}{
		{
			os:   "linux",
			arch: "x64",
		},
		{
			os:   "linux",
			arch: "arm64",
		},
		{
			os:   "macos",
			arch: "x64",
		},
		{
			os:   "macos",
			arch: "arm64",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.os, tt.arch), func(t *testing.T) {
			platformName := fmt.Sprintf("%s-%s", tt.os, tt.arch)
			expectedChecksum := mise.MiseChecksums[platformName]
			url := mise.FallbackDownloadURL(mise.MiseVersion, platformName)

			resp, err := retryablehttp.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code from %s: %d", url, resp.StatusCode)

			tempFile, err := os.CreateTemp("", "mise-test-*.tar.gz")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())
			defer tempFile.Close()

			hash := sha256.New()
			multiWriter := io.MultiWriter(tempFile, hash)
			_, err = io.Copy(multiWriter, resp.Body)
			require.NoError(t, err)

			calculatedChecksum := fmt.Sprintf("%x", hash.Sum(nil))
			require.Equal(t, expectedChecksum, calculatedChecksum, "SHA256 checksum mismatch for %s", platformName)
		})
	}
}
