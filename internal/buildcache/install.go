package buildcache

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v3/log"
	"github.com/hashicorp/go-retryablehttp"
)

const Version = "3.14.1"

// checksums come from the release's checksums.txt.
var checksums = map[string]string{
	"darwin_amd64": "8b0b67da9ee3409f9138bc8c8401202d3c625180e3d445ca854c2c46ad076029",
	"darwin_arm64": "afe2dbc666c1001b4ce45dee6e660936d4aa60512c3c978ff88e43b5d50d86a1",
	"linux_amd64":  "4a399c51e85ca45ac2dd984e16d5b16112edc139422fd2a8102db8721f14e432",
	"linux_arm64":  "af00aa9ad60c07cda06b1ecffa69c3f0ad98d220ce83948f93f841a59bb43f22",
}

// garFilesURL is the fallback when the host VM cache is unset or fails.
const garFilesURL = "https://artifactregistry.googleapis.com/v1/projects/ip-build-cache-prod/locations/us-central1/repositories/build-cache-cli-releases/files"

const (
	binaryName = "bitrise-build-cache"
	// Tarballs are ~8MB; this only stops a misbehaving server.
	maxTarballBytes = 64 << 20
)

type installer struct {
	logger       log.Logger
	client       *retryablehttp.Client
	hostCacheURL string
	garFilesURL  string
	checksums    map[string]string
}

func newInstaller(logger log.Logger, hostCacheURL string) installer {
	client := retryablehttp.NewClient()
	client.Logger = &log.HTTPLogAdaptor{Logger: logger}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	client.RetryMax = 2
	client.RetryWaitMax = 2 * time.Second
	client.HTTPClient.Timeout = time.Minute
	return installer{logger: logger, client: client, hostCacheURL: hostCacheURL, garFilesURL: garFilesURL, checksums: checksums}
}

func (i installer) install(ctx context.Context, dir string) (string, error) {
	bin := filepath.Join(dir, binaryName)
	if isExecutable(bin) {
		return bin, nil
	}

	platform := runtime.GOOS + "_" + runtime.GOARCH
	checksum, ok := i.checksums[platform]
	if !ok {
		return "", fmt.Errorf("no release for %s", platform)
	}

	var errs []error
	for n, url := range i.urls(platform) {
		fromHostCache := i.hostCacheURL != "" && n == 0
		if err := i.downloadBinary(ctx, url, checksum, bin, fromHostCache); err != nil {
			i.logger.Warnf("Downloading bitrise-build-cache from %s failed: %s", url, err)
			errs = append(errs, err)
			continue
		}
		i.logger.Infof("Installed bitrise-build-cache %s from %s", Version, url)
		return bin, nil
	}
	return "", errors.Join(errs...)
}

func (i installer) urls(platform string) []string {
	tarball := fmt.Sprintf("bitrise-build-cache_%s_%s.tar.gz", Version, platform)
	var urls []string
	if i.hostCacheURL != "" {
		urls = append(urls, strings.TrimSuffix(i.hostCacheURL, "/")+"/"+tarball)
	}
	gar := fmt.Sprintf("%s/bitrise-build-cache_%s.tar.gz:%s:%s:download?alt=media", i.garFilesURL, platform, Version, tarball)
	return append(urls, gar)
}

// downloadBinary verifies before extracting, so a partial or tampered download never reaches bin.
func (i installer) downloadBinary(ctx context.Context, url, checksum, bin string, fromHostCache bool) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if fromHostCache {
		// Same header the preboot scripts send to the host VM cache.
		req.Header.Set("Digest", "sha-256="+checksum)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status code %d", resp.StatusCode)
	}

	tarball, err := io.ReadAll(io.LimitReader(resp.Body, maxTarballBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	sum := sha256.Sum256(tarball)
	if got := hex.EncodeToString(sum[:]); got != checksum {
		return fmt.Errorf("checksum validation failed: expected %s, got %s", checksum, got)
	}

	return extractBinary(tarball, bin)
}

func extractBinary(tarball []byte, bin string) error {
	gz, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s not found in tarball", binaryName)
		}
		if err != nil {
			return fmt.Errorf("read tarball: %w", err)
		}
		if header.Typeflag == tar.TypeReg && path.Clean(header.Name) == binaryName {
			return writeExecutable(tr, bin)
		}
	}
}

func writeExecutable(r io.Reader, bin string) error {
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(bin), binaryName+"-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	if _, err := io.Copy(tmp, io.LimitReader(r, maxTarballBytes*4)); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("extract %s: %w", binaryName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return fmt.Errorf("chmod %s: %w", binaryName, err)
	}
	return os.Rename(tmp.Name(), bin)
}

func link(target, name string) error {
	if current, err := os.Readlink(name); err == nil && current == target {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	tmp := name + ".tmp"
	_ = os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, name)
}

func isExecutable(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}
