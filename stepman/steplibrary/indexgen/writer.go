package indexgen

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/fileutil"
)

// fileWriter abstracts the OS calls used by writeJSON, making them injectable
// for testing without affecting the fs.FS-based read path.
type fileWriter interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
}

type realFileWriter struct{}

// writer emits files under outputDir and tracks file count + byte count for Stats.
type writer struct {
	outputDir string
	fw        fileWriter
	fm        fileutil.FileManager
	fileCount int
	byteCount int64
}

func (w *writer) writeJSON(relPath string, v any) error {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	full := filepath.Join(w.outputDir, relPath)
	// Files we author get owner-only perms (no group/other needed).
	if err := w.fw.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		return err
	}
	if err := w.fw.WriteFile(full, bytes, 0o600); err != nil {
		return err
	}
	w.fileCount++
	w.byteCount += int64(len(bytes))
	return nil
}

func (w *writer) copyFileFromFS(srcFS fs.FS, srcPath, relDst string) error {
	dst := filepath.Join(w.outputDir, relDst)
	// CopyFileFS opens dst directly and does not create parent dirs, so create
	// the containing dir ourselves (owner-only; the copied file keeps its source
	// perms, which CopyFileFS preserves).
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	if err := w.fm.CopyFileFS(srcFS, srcPath, dst, &fileutil.CopyOptions{Overwrite: true}); err != nil {
		return err
	}
	info, err := fs.Stat(srcFS, srcPath)
	if err != nil {
		return err
	}
	w.fileCount++
	w.byteCount += info.Size()
	return nil
}

func (realFileWriter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (realFileWriter) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
