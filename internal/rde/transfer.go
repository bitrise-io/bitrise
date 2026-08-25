package rde

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// gcsHTTPTimeout caps each cloud-storage transfer leg (PUT during upload,
// GET during download). Sessions can hold large workspaces, so 10 minutes is
// a generous ceiling for either leg.
const gcsHTTPTimeout = 10 * time.Minute

// UploadFile uploads a local file or directory to a session: tars the
// source, gzips it, PUTs it to the signed URL the backend returns, then
// calls complete-file-upload to trigger extraction at destFolder.
func (s *Service) UploadFile(ctx context.Context, workspaceID, sessionID, sourcePath, destFolder string) error {
	if s.client == nil {
		return errClient()
	}
	if sourcePath == "" {
		return fmt.Errorf("source path is required")
	}
	if destFolder == "" {
		return fmt.Errorf("destination folder is required")
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	start, err := s.client.SessionStartFileUpload(ctx, workspaceID, sessionID, rdeapi.StartFileUploadRequest{
		DestinationFolder: destFolder,
	})
	if err != nil {
		return fmt.Errorf("start file upload: %w", err)
	}

	archive, err := createTarGz(sourcePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}

	if err := putToSignedURL(ctx, start.SignedURL, archive); err != nil {
		return fmt.Errorf("upload to cloud storage: %w", err)
	}

	if err := s.client.SessionCompleteFileUpload(ctx, workspaceID, sessionID, rdeapi.CompleteFileUploadRequest{
		UploadID:          start.UploadID,
		DestinationFolder: destFolder,
	}); err != nil {
		return fmt.Errorf("complete file upload: %w", err)
	}
	return nil
}

// DownloadFile downloads remote sourcePath from the session into localDest.
// When onlyContents is true and the remote path is a directory, only the
// directory's contents are extracted (not the directory itself).
func (s *Service) DownloadFile(ctx context.Context, workspaceID, sessionID, sourcePath, localDest string, onlyContents bool) error {
	if s.client == nil {
		return errClient()
	}
	if localDest == "" {
		return fmt.Errorf("local destination is required")
	}

	resp, err := s.client.SessionDownloadFile(ctx, workspaceID, sessionID, rdeapi.DownloadFileRequest{
		SourcePath:           sourcePath,
		OnlyContentsOfFolder: onlyContents,
	})
	if err != nil {
		return fmt.Errorf("request download: %w", err)
	}

	archive, err := getFromSignedURL(ctx, resp.SignedURL)
	if err != nil {
		return fmt.Errorf("download from cloud storage: %w", err)
	}

	if err := extractTarGz(archive, localDest); err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}
	return nil
}

// createTarGz walks sourcePath and returns a gzipped tar archive of its
// contents. A directory becomes an archive of its tree (relative paths);
// a single file becomes an archive containing just that file.
func createTarGz(sourcePath string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat source: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	if info.IsDir() {
		baseDir = sourcePath
	}

	addEntry := func(path string, fi os.FileInfo, name string) error {
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return fmt.Errorf("file info header: %w", err)
		}
		header.Name = name
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header: %w", err)
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(path) // path comes from filepath.WalkDir under sourcePath
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close() //nolint:errcheck // copy error takes precedence
		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}
		return nil
	}

	if info.IsDir() {
		walkErr := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			fi, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			return addEntry(path, fi, rel)
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walk directory: %w", walkErr)
		}
	} else {
		if err := addEntry(sourcePath, info, filepath.Base(sourcePath)); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip: %w", err)
	}
	return buf.Bytes(), nil
}

func putToSignedURL(ctx context.Context, signedURL string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/gzip")

	client := http.Client{Timeout: gcsHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // body fully consumed below; close error is non-actionable

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func getFromSignedURL(ctx context.Context, signedURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	client := http.Client{Timeout: gcsHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // body fully consumed below; close error is non-actionable
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

// extractTarGz extracts a gzipped tar archive into destDir. Symlinks and
// other non-regular non-dir entries are skipped silently. Zip-slip is
// prevented by resolving each entry against destDir and rejecting any whose
// relative path escapes it — one bad entry aborts the whole extraction (no
// partial recovery).
func extractTarGz(data []byte, destDir string) error {
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}

	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close() //nolint:errcheck // closing a reader, fine to ignore

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		target := filepath.Join(absDest, header.Name) // zip-slip guarded immediately below
		// Zip-slip guard: refuse entries whose joined path escapes destDir.
		if rel, relErr := filepath.Rel(absDest, target); relErr != nil || rel == ".." || filepath.IsAbs(rel) || (len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive entry %q would escape destination", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o750); err != nil {
				return fmt.Errorf("create dir: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return fmt.Errorf("create parent dir: %w", err)
			}
			// Mode comes from a trusted in-process tar header parse; mask to
			// the standard 9 permission bits before casting to a FileMode
			// without losing the rwxrwxrwx bits we actually care about.
			mode := os.FileMode(header.Mode & 0o777)
			if err := writeTarFile(tr, target, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeTarFile copies the current tar reader entry into a regular file at
// target with the given mode. Split out of extractTarGz so the defer in
// the loop body doesn't leak file descriptors across many entries.
func writeTarFile(tr *tar.Reader, target string, mode os.FileMode) error {
	f, err := os.Create(target) // target is the verified absolute path inside destDir
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()                           //nolint:errcheck // copy error takes precedence
	if _, err := io.Copy(f, tr); err != nil { // archive size is bounded by what GCS served us
		return fmt.Errorf("write file: %w", err)
	}
	if err := os.Chmod(target, mode); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	return nil
}
