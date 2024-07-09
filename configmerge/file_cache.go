package configmerge

import (
	"io"
	"os"
	"path/filepath"

	logV2 "github.com/bitrise-io/go-utils/v2/log"
)

type fileCache struct {
	logger   logV2.Logger
	cacheDir string
}

func NewFileCache(cacheDir string, logger logV2.Logger) FileCache {
	return fileCache{
		logger:   logger,
		cacheDir: cacheDir,
	}
}

func (c fileCache) GetFileContent(key string) ([]byte, error) {
	pth := filepath.Join(c.cacheDir, key)
	f, err := os.Open(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			c.logger.Warnf("Failed to close cache file: %s", err)
		}
	}()
	return io.ReadAll(f)
}

func (c fileCache) SetFileContent(key string, content []byte) error {
	pth := filepath.Join(c.cacheDir, key)
	f, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			c.logger.Warnf("Failed to close cache file: %s", err)
		}
	}()
	if _, err := f.Write(content); err != nil {
		return err
	}
	return nil
}
