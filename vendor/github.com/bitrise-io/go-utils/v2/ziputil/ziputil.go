package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// ZipManager provides zip and unzip operations using Go stdlib.
type ZipManager struct {
	pathChecker pathutil.PathChecker
	osProxy     OsProxy
}

// NewZipManager creates a ZipManager backed by the real OS.
func NewZipManager(pathChecker pathutil.PathChecker) *ZipManager {
	return &ZipManager{pathChecker: pathChecker, osProxy: RealOS{}}
}

// ZipDir zips sourceDirPth into destinationZipPth.
// isContentOnly=false: archive contains the directory under its own basename (e.g. "mydir/file.txt").
// isContentOnly=true: archive contains the directory's contents directly (e.g. "file.txt").
// Directory modification times are preserved. Symlinks are stored as symlinks (not followed).
// If destinationZipPth has no file extension, ".zip" is appended.
func (z *ZipManager) ZipDir(sourceDirPth, destinationZipPth string, isContentOnly bool) error {
	if exist, err := z.pathChecker.IsDirExists(sourceDirPth); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("directory (%s) does not exist", sourceDirPth)
	}

	baseDir := filepath.Dir(sourceDirPth)
	if isContentOnly {
		baseDir = sourceDirPth
	}

	return z.createZipFromDir(destinationZipPth, sourceDirPth, baseDir)
}

// ZipDirs zips multiple directories into a single archive, each under its own basename.
// When two entries in sourceDirPths share the same basename (e.g. "/a/shared" and "/b/shared"),
// their contents are merged: files unique to either directory are preserved, and conflicting
// files (same relative path) resolve in favour of the last directory in the list.
// If destinationZipPth has no file extension, ".zip" is appended.
func (z *ZipManager) ZipDirs(sourceDirPths []string, destinationZipPth string) error {
	for _, path := range sourceDirPths {
		if exist, err := z.pathChecker.IsDirExists(path); err != nil {
			return err
		} else if !exist {
			return fmt.Errorf("directory (%s) does not exist", path)
		}
	}

	return z.createZipFile(destinationZipPth, func(zw *zip.Writer) error {
		for _, sourceDirPth := range sourceDirPths {
			if err := z.addDirToZip(zw, sourceDirPth, filepath.Dir(sourceDirPth)); err != nil {
				return err
			}
		}
		return nil
	})
}

// ZipFile zips a single file into destinationZipPth.
// If destinationZipPth has no file extension, ".zip" is appended.
func (z *ZipManager) ZipFile(sourceFilePth, destinationZipPth string) error {
	return z.ZipFiles([]string{sourceFilePth}, destinationZipPth)
}

// ZipFiles zips multiple files into destinationZipPth without preserving directory structure.
// Each file is stored under its base name only. Symlinks are stored as symlinks (not followed).
// If two source files share the same base name, an error is returned before the destination
// file is created. If destinationZipPth has no file extension, ".zip" is appended.
func (z *ZipManager) ZipFiles(sourceFilePths []string, destinationZipPth string) error {
	seen := make(map[string]bool)
	for _, path := range sourceFilePths {
		if exist, err := z.pathChecker.IsPathExists(path); err != nil {
			return err
		} else if !exist {
			return fmt.Errorf("file (%s) does not exist", path)
		}
		baseName := filepath.Base(path)
		if seen[baseName] {
			return fmt.Errorf("duplicate file name %q: files with the same base name cannot be zipped together", baseName)
		}
		seen[baseName] = true
	}

	return z.createZipFile(destinationZipPth, func(zw *zip.Writer) error {
		for _, filePath := range sourceFilePths {
			if err := z.addFileToZip(zw, filePath, filepath.Base(filePath)); err != nil {
				return err
			}
		}
		return nil
	})
}

// UnZip extracts the zip archive at zipPth into intoDir, restoring permissions and symlinks.
// Entries with path-traversal components (e.g. "../../escape") are rejected and an error is
// returned immediately; no further entries are extracted after a traversal is detected.
// Rejection is by substring: any entry whose name contains ".." anywhere is refused, including
// otherwise-legitimate names like "foo..bar". This is broader than strict path-component
// traversal, but is required for the repo's static analysis to recognise the sanitizer.
func (z *ZipManager) UnZip(zipPth, intoDir string) error {
	r, err := zip.OpenReader(zipPth)
	if err != nil {
		return err
	}
	defer r.Close() //nolint:errcheck

	cleanDest := filepath.Clean(intoDir)
	// Ensure intoDir exists so EvalSymlinks can resolve it before processing any entry.
	if err := z.osProxy.MkdirAll(cleanDest, 0755); err != nil {
		return err
	}
	// Resolve once per call; EvalSymlinks is called per-entry otherwise (O(n) syscalls).
	// Both cleanDest and realDest are passed to extractEntry to avoid re-computing them.
	realDest, err := z.osProxy.EvalSymlinks(cleanDest)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		// CodeQL's go/zipslip recognises strings.Contains(f.Name, "..") as the canonical
		// taint sanitizer at the loop source. Must be standalone here — a compound && check
		// leaves a reachable path where strings.Contains is true but execution continues,
		// which CodeQL flags as unsanitized. Do not extract into a helper.
		// Tradeoff: any entry whose name contains ".." as a substring is rejected, including
		// names like "module..framework". No such names appear in iOS/Xcode artifact formats.
		if strings.Contains(f.Name, "..") {
			return fmt.Errorf("illegal path in zip entry: %s", f.Name)
		}
		if err := z.extractEntry(f, cleanDest, realDest); err != nil {
			return err
		}
	}
	return nil
}

func (z *ZipManager) createZipFromDir(destinationZipPth, sourceDirPth, baseDir string) error {
	return z.createZipFile(destinationZipPth, func(zw *zip.Writer) error {
		return z.addDirToZip(zw, sourceDirPth, baseDir)
	})
}

// createZipFile builds a zip archive at destinationZipPth, delegating entry creation to
// addEntries. If destinationZipPth has no file extension, ".zip" is appended (an existing
// extension is left unchanged). It writes to a temporary file in the destination directory and
// renames it over the destination only after writing succeeds, so a failed write leaves a
// pre-existing destination untouched.
func (z *ZipManager) createZipFile(destinationZipPth string, addEntries func(zw *zip.Writer) error) (retErr error) {
	if filepath.Ext(destinationZipPth) == "" {
		destinationZipPth += ".zip"
	}
	tmpPath := destinationZipPth + ".tmp"

	dest, err := z.osProxy.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			z.osProxy.Remove(tmpPath) //nolint:errcheck
		}
	}()

	zw := zip.NewWriter(dest)
	if err := addEntries(zw); err != nil {
		dest.Close() //nolint:errcheck
		return err
	}
	if err := zw.Close(); err != nil {
		dest.Close() //nolint:errcheck
		return err
	}
	if err := dest.Close(); err != nil {
		return err
	}

	return z.osProxy.Rename(tmpPath, destinationZipPth)
}

func (z *ZipManager) addDirToZip(zw *zip.Writer, sourceDirPth, baseDir string) error {
	return filepath.WalkDir(sourceDirPth, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return z.addSymlinkToZip(zw, path, relPath)
		}

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			hdr := &zip.FileHeader{
				Name:     relPath + "/",
				Method:   zip.Store,
				Modified: info.ModTime(),
			}
			hdr.SetMode(info.Mode())
			_, err = zw.CreateHeader(hdr)
			return err
		}

		return z.addFileToZip(zw, path, relPath)
	})
}

func (z *ZipManager) addSymlinkToZip(zw *zip.Writer, path, name string) error {
	target, err := z.osProxy.Readlink(path)
	if err != nil {
		return err
	}

	hdr := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}
	hdr.SetMode(os.ModeSymlink | 0777)

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(target))
	return err
}

func (z *ZipManager) addFileToZip(zw *zip.Writer, path, name string) error {
	info, err := z.osProxy.Lstat(path)
	if err != nil {
		return err
	}

	// If the path is a symlink, store it as a symlink rather than following it.
	if info.Mode()&os.ModeSymlink != 0 {
		return z.addSymlinkToZip(zw, path, name)
	}

	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Method = zip.Deflate

	src, err := z.osProxy.Open(path)
	if err != nil {
		return err
	}
	defer src.Close() //nolint:errcheck

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}

// extractEntry extracts a single zip entry into cleanDest.
// cleanDest must be filepath.Clean(intoDir); realDest must be EvalSymlinks(cleanDest).
// Both are pre-computed by UnZip to avoid redundant syscalls across entries.
func (z *ZipManager) extractEntry(f *zip.File, cleanDest, realDest string) error {
	// Duplicates UnZip's strings.Contains pre-filter on purpose: it keeps the ".." sanitizer in
	// the same function as the file-system sinks below, which is what CodeQL's go/zipslip query
	// needs to treat them as sanitized. It is also genuine defense-in-depth. Do not remove it
	// just because UnZip already rejects these names before calling extractEntry.
	if strings.Contains(f.Name, "..") {
		return fmt.Errorf("illegal path in zip entry: %s", f.Name)
	}
	sep := string(os.PathSeparator)
	destPath := filepath.Clean(filepath.Join(cleanDest, f.Name))
	if destPath != cleanDest && !strings.HasPrefix(destPath, cleanDest+sep) {
		return fmt.Errorf("illegal path in zip entry: %s", f.Name)
	}

	if f.Mode()&os.ModeSymlink != 0 {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close() //nolint:errcheck

		target, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		targetStr := string(target)
		if filepath.IsAbs(targetStr) {
			return fmt.Errorf("symlink target %q is absolute", targetStr)
		}
		resolvedTarget := filepath.Clean(filepath.Join(filepath.Dir(destPath), targetStr))
		if resolvedTarget != cleanDest && !strings.HasPrefix(resolvedTarget, cleanDest+sep) {
			return fmt.Errorf("symlink target %q escapes extraction directory", targetStr)
		}
		if err := z.osProxy.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		// Resolve pre-existing symlinks in the parent path (go/unsafe-unzip-symlink): a
		// previously extracted symlink could make the parent resolve outside cleanDest even
		// though the syntactic path appears inside it.
		realParent, err := z.osProxy.EvalSymlinks(filepath.Dir(destPath))
		if err != nil {
			return err
		}
		if realParent != realDest && !strings.HasPrefix(realParent, realDest+sep) {
			return fmt.Errorf("symlink parent %q escapes extraction directory", filepath.Dir(destPath))
		}
		return z.osProxy.Symlink(targetStr, destPath)
	}

	if f.FileInfo().IsDir() {
		return z.osProxy.MkdirAll(destPath, f.Mode().Perm())
	}

	if err := z.osProxy.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	// Same chained-symlink check for regular files.
	realParent, err := z.osProxy.EvalSymlinks(filepath.Dir(destPath))
	if err != nil {
		return err
	}
	if realParent != realDest && !strings.HasPrefix(realParent, realDest+sep) {
		return fmt.Errorf("file parent %q escapes extraction directory", filepath.Dir(destPath))
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close() //nolint:errcheck

	dest, err := z.osProxy.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode().Perm())
	if err != nil {
		return err
	}
	defer dest.Close() //nolint:errcheck

	_, err = io.Copy(dest, rc)
	return err
}
