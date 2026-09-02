package rde

import (
	"archive/tar"
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
// a generous ceiling for either leg. It is an http.Client timeout, so on
// download it covers extraction too — the archive is extracted as it arrives.
// Extraction is disk-bound and short next to the network leg, so the same
// ceiling still applies comfortably.
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

	// Archived BEFORE the upload is registered: a failure here would otherwise
	// leave an upload record and blob the backend never sees completed, and
	// archiving a large tree after the fact eats into the signed URL's
	// validity window.
	archive, size, err := createTarGzFile(sourcePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer func() {
		_ = archive.Close()
		_ = os.Remove(archive.Name())
	}()

	start, err := s.client.SessionStartFileUpload(ctx, workspaceID, sessionID, rdeapi.StartFileUploadRequest{
		DestinationFolder: destFolder,
	})
	if err != nil {
		return fmt.Errorf("start file upload: %w", err)
	}

	if err := putToSignedURL(ctx, start.SignedURL, archive, size); err != nil {
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

	body, err := openSignedURL(ctx, resp.SignedURL)
	if err != nil {
		return fmt.Errorf("download from cloud storage: %w", err)
	}
	defer body.Close() //nolint:errcheck // extract error takes precedence

	if err := extractTarGz(body, localDest); err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}
	return nil
}

// createTarGzFile writes a gzipped tar of sourcePath to a temp file and returns
// it rewound, along with its size. Going through a file rather than returning a
// []byte keeps a multi-GB session workspace off the heap.
//
// The caller owns the returned file: close it and remove it by Name().
func createTarGzFile(sourcePath string) (*os.File, int64, error) {
	f, err := os.CreateTemp("", "bitrise-rde-upload-*.tar.gz")
	if err != nil {
		return nil, 0, fmt.Errorf("create temp archive: %w", err)
	}
	discard := func(err error) (*os.File, int64, error) {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, 0, err
	}

	if err := writeTarGz(f, sourcePath); err != nil {
		return discard(err)
	}
	fi, err := f.Stat()
	if err != nil {
		return discard(fmt.Errorf("stat archive: %w", err))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return discard(fmt.Errorf("rewind archive: %w", err))
	}
	return f, fi.Size(), nil
}

// writeTarGz streams a gzipped tar of sourcePath into w. A directory becomes an
// archive of its tree (relative paths); a single file becomes an archive
// containing just that file.
func writeTarGz(w io.Writer, sourcePath string) error {
	gw := gzip.NewWriter(w)
	tw := tar.NewWriter(gw)

	info, err := os.Stat(sourcePath) // follows a symlinked source
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	if info.IsDir() {
		// os.Stat followed the link, but filepath.WalkDir lstats its root and
		// would emit the link itself instead of descending, so walk whatever
		// the source resolves to.
		resolved, resolveErr := filepath.EvalSymlinks(sourcePath)
		if resolveErr != nil {
			return fmt.Errorf("resolve source: %w", resolveErr)
		}
		sourcePath = resolved
		baseDir = resolved
	}

	addEntry := func(path string, fi os.FileInfo, name string) error {
		// A socket has no tar representation at all: tar.FileInfoHeader rejects
		// it outright ("archive/tar: sockets not supported"), which would abort
		// the whole upload over one stray file. Devices and FIFOs do have one,
		// but recreating them on the session needs privileges and they carry
		// nothing a workspace transfer needs.
		if fi.Mode()&(os.ModeSocket|os.ModeDevice|os.ModeCharDevice|os.ModeNamedPipe) != 0 {
			return nil
		}

		link := ""
		if fi.Mode()&os.ModeSymlink != 0 {
			// WalkDir's DirEntry.Info is an Lstat, so fi describes the link
			// itself. Without its target the entry is a symlink pointing at
			// nothing, and extraction on the session fails on it.
			target, readErr := os.Readlink(path)
			if readErr != nil {
				return fmt.Errorf("read symlink: %w", readErr)
			}
			link = target
		}

		header, err := tar.FileInfoHeader(fi, link)
		if err != nil {
			return fmt.Errorf("file info header: %w", err)
		}
		header.Name = filepath.ToSlash(name) // tar names are slash-separated on every platform
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
			return fmt.Errorf("walk directory: %w", walkErr)
		}
	} else if err := addEntry(sourcePath, info, filepath.Base(sourcePath)); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("close gzip: %w", err)
	}
	return nil
}

func putToSignedURL(ctx context.Context, signedURL string, body io.ReadSeeker, size int64) error {
	// NopCloser keeps ownership of the file with the caller: net/http closes a
	// body that implements io.Closer, which would break both the caller's
	// cleanup and the GetBody replay below.
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, io.NopCloser(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/gzip")
	// Declared explicitly: for a body that isn't one of the types net/http can
	// measure itself, the request would otherwise go out with chunked transfer
	// encoding, which signed-URL uploads reject.
	req.ContentLength = size
	// net/http only derives GetBody for in-memory body types. Without it, a
	// 307/308 (an S3-compatible endpoint redirecting to the bucket's region) is
	// not followed and not reported either — Do hands back the 3xx itself.
	req.GetBody = func() (io.ReadCloser, error) {
		if _, err := body.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("rewind body: %w", err)
		}
		return io.NopCloser(body), nil
	}

	client := http.Client{Timeout: gcsHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // body fully consumed below; close error is non-actionable

	// Anything outside 2xx, a redirect that was handed back unfollowed
	// included: treating only >= 400 as failure reports success for an upload
	// that never landed.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(msg))
	}
	return nil
}

// openSignedURL issues the GET and hands back the response body unread, so the
// archive can be extracted as it arrives instead of being buffered whole. The
// caller owns the returned reader and must close it.
func openSignedURL(ctx context.Context, signedURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	client := http.Client{Timeout: gcsHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(msg))
	}
	return resp.Body, nil
}

// extractTarGz extracts a gzipped tar stream into destDir. Symlinks and
// other non-regular non-dir entries are skipped silently. Zip-slip is
// prevented by resolving each entry against destDir and rejecting any whose
// relative path escapes it — one bad entry aborts the whole extraction (no
// partial recovery). Aborting leaves the entries already written in destDir:
// extraction is interleaved with the download, so a truncated transfer leaves
// a partial tree behind rather than an untouched destination.
func extractTarGz(r io.Reader, destDir string) error {
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}

	gr, err := gzip.NewReader(r)
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
