package fileutil

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyOptions configures a [FileManager.CopyFile] operation.
// A nil pointer means default behavior.
type CopyOptions struct {
	Overwrite bool // Overwrite replaces the destination file if it already exists.
}

// FileManager ...
type FileManager interface {
	Open(path string) (*os.File, error)
	OpenReaderIfExists(path string) (io.Reader, error)
	ReadDirEntryNames(path string) ([]string, error)
	Remove(path string) error
	RemoveAll(path string) error
	Write(path string, value string, perm os.FileMode) error
	WriteBytes(path string, value []byte) error
	FileSizeInBytes(pth string) (int64, error)
	CopyFile(src, dst string, opts *CopyOptions) error
	CopyDir(src, dst string, opts *CopyOptions) error
	Lstat(path string) (os.FileInfo, error)
	LastNLines(s string, n int) string
}

type fileManager struct {
	osProxy OsProxy
}

// NewFileManager ...
func NewFileManager() FileManager {
	return fileManager{osProxy: RealOS{}}
}

// ReadDirEntryNames reads the named directory using os.ReadDir and returns the dir entries' names.
func (fileManager) ReadDirEntryNames(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

// Open ...
func (fileManager) Open(path string) (*os.File, error) {
	return os.Open(path)
}

// OpenReaderIfExists opens the named file using os.Open and returns an io.Reader.
// An ErrNotExist error is absorbed and the returned io.Reader will be nil,
// other errors from os.Open are returned as is.
func (fileManager) OpenReaderIfExists(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Remove ...
func (fileManager) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll ...
func (fileManager) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Write ...
func (f fileManager) Write(path string, value string, mode os.FileMode) error {
	if err := f.ensureSavePath(path); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(value), mode); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

func (fileManager) ensureSavePath(savePath string) error {
	dirPath := filepath.Dir(savePath)
	return os.MkdirAll(dirPath, 0700)
}

// WriteBytes ...
func (f fileManager) WriteBytes(path string, value []byte) error {
	return os.WriteFile(path, value, 0600)
}

// FileSizeInBytes checks if the provided path exists and return with the file size (bytes) using os.Lstat.
func (fileManager) FileSizeInBytes(pth string) (int64, error) {
	if pth == "" {
		return 0, errors.New("No path provided")
	}
	fileInf, err := os.Stat(pth)
	if err != nil {
		return 0, err
	}

	return fileInf.Size(), nil
}

// Lstat implements FileManager by delegating to os.Lstat via the osProxy.
func (fm fileManager) Lstat(path string) (os.FileInfo, error) {
	return fm.osProxy.Lstat(path)
}

// LastNLines returns the last n lines of the given string s.
func (fileManager) LastNLines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	// normalize CRLF to LF if needed
	if strings.Contains(s, "\r\n") {
		s = strings.ReplaceAll(s, "\r\n", "\n")
	}

	// skip trailing newlines so we don't count empty trailing lines
	i := len(s) - 1
	for i >= 0 && s[i] == '\n' {
		i--
	}
	if i < 0 {
		return "" // string was all newlines
	}

	// scan backward counting '\n' occurrences
	count := 0
	for ; i >= 0; i-- {
		if s[i] == '\n' {
			count++
			if count == n {
				// substring after this newline is the last n lines
				start := i + 1
				res := s[start:]
				// trim trailing whitespace (spaces, tabs, newlines, etc.)
				return strings.TrimRightFunc(res, func(r rune) bool {
					return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
				})
			}
		}
	}

	// fewer than n newlines => return whole string (trim trailing whitespace)
	return strings.TrimRightFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
	})
}
