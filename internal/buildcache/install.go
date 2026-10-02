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

// garFilesURL is the fallback when the host VM cache is unset or fails.
const garFilesURL = "https://artifactregistry.googleapis.com/v1/projects/ip-build-cache-prod/locations/us-central1/repositories/build-cache-cli-releases/files"

const (
	binaryName = "bitrise-build-cache"
	// Tarballs are ~8MB; this only stops a misbehaving server.
	maxTarballBytes = 64 << 20
	// Per source, retries included, so a blackholed host cache still leaves time for GAR.
	attemptTimeout = 30 * time.Second
)

type installer struct {
	logger       log.Logger
	client       *retryablehttp.Client
	hostCacheURL string
	garFilesURL  string
	version      string
	checksum     string
	attempt      time.Duration
}

func newInstaller(logger log.Logger, hostCacheURL, version, checksum string) installer {
	client := retryablehttp.NewClient()
	client.Logger = &log.HTTPLogAdaptor{Logger: logger}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	client.RetryMax = 2
	client.RetryWaitMax = 2 * time.Second
	client.HTTPClient.Timeout = time.Minute
	return installer{logger: logger, client: client, hostCacheURL: hostCacheURL, garFilesURL: garFilesURL, version: version, checksum: checksum, attempt: attemptTimeout}
}

func (i installer) install(ctx context.Context, dir string) (string, error) {
	bin := filepath.Join(dir, binaryName)
	if isExecutable(bin) {
		return bin, nil
	}

	var errs []error
	for n, url := range i.urls(runtime.GOOS + "_" + runtime.GOARCH) {
		fromHostCache := i.hostCacheURL != "" && n == 0
		attemptCtx, cancel := context.WithTimeout(ctx, i.attempt)
		err := i.downloadBinary(attemptCtx, url, bin, fromHostCache)
		cancel()
		if err != nil {
			i.logger.Warnf("Downloading bitrise-build-cache from %s failed: %s", url, err)
			errs = append(errs, err)
			continue
		}
		i.logger.Infof("Installed bitrise-build-cache %s from %s", i.version, url)
		return bin, nil
	}
	return "", errors.Join(errs...)
}

func (i installer) urls(platform string) []string {
	tarball := fmt.Sprintf("bitrise-build-cache_%s_%s.tar.gz", i.version, platform)
	var urls []string
	if i.hostCacheURL != "" {
		urls = append(urls, strings.TrimSuffix(i.hostCacheURL, "/")+"/"+tarball)
	}
	gar := fmt.Sprintf("%s/bitrise-build-cache_%s.tar.gz:%s:%s:download?alt=media", i.garFilesURL, platform, i.version, tarball)
	return append(urls, gar)
}

// downloadBinary verifies before extracting, so a partial or tampered download never reaches bin.
func (i installer) downloadBinary(ctx context.Context, url, bin string, fromHostCache bool) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if fromHostCache {
		// Same header the preboot scripts send to the host VM cache.
		req.Header.Set("Digest", "sha-256="+i.checksum)
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
	if got := hex.EncodeToString(sum[:]); got != i.checksum {
		return fmt.Errorf("checksum validation failed: expected %s, got %s", i.checksum, got)
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
	tmp := fmt.Sprintf("%s.tmp-%d-%d", name, os.Getpid(), time.Now().UnixNano())
	if err := os.Symlink(target, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, name); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func isExecutable(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}
