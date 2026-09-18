package rde

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTarGzFile_DirectoryTree(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "top.txt"), "top")
	if err := os.MkdirAll(filepath.Join(src, "nested", "deep"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(src, "nested", "deep", "leaf.txt"), "leaf")

	entries := archiveEntries(t, src)

	if got, ok := entries["top.txt"]; !ok || got.body != "top" {
		t.Errorf("top.txt = %+v (present=%v), want body %q", got, ok, "top")
	}
	// Tar names are slash-separated on every platform, so filepath.Rel output
	// must be converted — otherwise a Windows upload extracts on the session as
	// one file literally named `nested\deep\leaf.txt`.
	if got, ok := entries["nested/deep/leaf.txt"]; !ok || got.body != "leaf" {
		t.Errorf("nested/deep/leaf.txt = %+v (present=%v), want body %q", got, ok, "leaf")
	}
	for name := range entries {
		if strings.Contains(name, `\`) {
			t.Errorf("entry name %q contains a backslash; tar names must be slash-separated", name)
		}
	}
}

func TestCreateTarGzFile_SingleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "only.txt")
	writeFile(t, path, "content")

	entries := archiveEntries(t, path)

	if len(entries) != 1 {
		t.Fatalf("entries = %v, want exactly one", entries)
	}
	got, ok := entries["only.txt"]
	if !ok {
		t.Fatalf("entries = %v, want a single entry named only.txt", entries)
	}
	if got.body != "content" {
		t.Errorf("only.txt body = %q, want %q", got.body, "content")
	}
}

func TestExtractTarGz_RejectsEscapingEntries(t *testing.T) {
	for _, name := range []string{"../evil", "..", "a/../../evil", "../../../../etc/passwd"} {
		t.Run(name, func(t *testing.T) {
			dest := t.TempDir()
			err := extractTarGz(bytes.NewReader(tarGzWithEntry(t, name, "pwned")), dest)
			if err == nil {
				t.Fatalf("extractTarGz accepted escaping entry %q", name)
			}
			if !strings.Contains(err.Error(), "escape destination") {
				t.Errorf("error = %q, want it to mention escaping the destination", err)
			}
		})
	}
}

// An absolute entry name is not an escape: filepath.Join absorbs the leading
// separator, so the entry lands inside destDir rather than at the archive's
// path. Asserted so the containment is deliberate rather than incidental.
func TestExtractTarGz_ContainsAbsoluteEntryNames(t *testing.T) {
	dest := t.TempDir()
	if err := extractTarGz(bytes.NewReader(tarGzWithEntry(t, "/etc/passwd", "contained")), dest); err != nil {
		t.Fatalf("extractTarGz: %v", err)
	}
	if got := readFile(t, filepath.Join(dest, "etc", "passwd")); got != "contained" {
		t.Errorf("extracted body = %q, want %q", got, "contained")
	}
}

func TestCreateTarGzFile_ExtractRoundTrip(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "alpha")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	script := filepath.Join(src, "sub", "b.sh")
	writeFile(t, script, "#!/bin/sh\n")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	archive, _, err := createTarGzFile(src)
	if err != nil {
		t.Fatalf("createTarGzFile: %v", err)
	}
	defer func() {
		_ = archive.Close()
		_ = os.Remove(archive.Name())
	}()

	dest := t.TempDir()
	if err := extractTarGz(archive, dest); err != nil {
		t.Fatalf("extractTarGz: %v", err)
	}

	if got := readFile(t, filepath.Join(dest, "a.txt")); got != "alpha" {
		t.Errorf("a.txt = %q, want %q", got, "alpha")
	}
	extracted := filepath.Join(dest, "sub", "b.sh")
	if got := readFile(t, extracted); got != "#!/bin/sh\n" {
		t.Errorf("sub/b.sh = %q, want %q", got, "#!/bin/sh\n")
	}
	fi, err := os.Stat(extracted)
	if err != nil {
		t.Fatalf("stat extracted script: %v", err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("sub/b.sh mode = %v, want %v", fi.Mode().Perm(), os.FileMode(0o755))
	}
}

func TestCreateTarGzFile_MissingSourceLeavesNoTempFile(t *testing.T) {
	// os.CreateTemp("") and os.TempDir() both resolve $TMPDIR, so redirecting it
	// keeps the count from seeing another test run's archives. Windows reads
	// TMP/TEMP instead, where this is merely no worse than the shared dir.
	t.Setenv("TMPDIR", t.TempDir())

	before := tempFileCount(t)

	if _, _, err := createTarGzFile(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("createTarGzFile succeeded for a missing source, want error")
	}

	if after := tempFileCount(t); after != before {
		t.Errorf("temp archive count went %d -> %d; the failed archive was not cleaned up", before, after)
	}
}

// TestPutToSignedURL_FollowsTemporaryRedirect pins the GetBody wiring. net/http
// derives GetBody only for in-memory body types; with an *os.File body and none
// set, a 307/308 is neither followed nor reported — Do hands back the 3xx, and a
// ">= 400" success check then calls that a successful upload.
func TestPutToSignedURL_FollowsTemporaryRedirect(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var landed []byte
			mux := http.NewServeMux()
			mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read redirected body: %v", err)
				}
				landed = body
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()
			mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, srv.URL+"/final", status)
			})

			body, size := seekableBody(t, "payload-bytes")
			if err := putToSignedURL(context.Background(), srv.URL+"/start", body, size); err != nil {
				t.Fatalf("putToSignedURL: %v", err)
			}
			if string(landed) != "payload-bytes" {
				t.Errorf("redirect target received %q, want %q", landed, "payload-bytes")
			}
		})
	}
}

func TestPutToSignedURL_StatusHandling(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantError bool
	}{
		{name: "200 OK", status: http.StatusOK},
		{name: "201 Created", status: http.StatusCreated},
		// A 3xx handed back unfollowed means nothing was uploaded, so it must
		// not read as success.
		{name: "304 Not Modified", status: http.StatusNotModified, wantError: true},
		{name: "403 Forbidden", status: http.StatusForbidden, wantError: true},
		{name: "500 Internal Server Error", status: http.StatusInternalServerError, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != "application/gzip" {
					t.Errorf("Content-Type = %q, want application/gzip", r.Header.Get("Content-Type"))
				}
				if r.TransferEncoding != nil {
					t.Errorf("TransferEncoding = %v, want none (ContentLength must be declared)", r.TransferEncoding)
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			body, size := seekableBody(t, "payload")
			err := putToSignedURL(context.Background(), srv.URL, body, size)
			if tt.wantError && err == nil {
				t.Fatalf("putToSignedURL returned nil for status %d, want error", tt.status)
			}
			if !tt.wantError && err != nil {
				t.Fatalf("putToSignedURL: %v", err)
			}
		})
	}
}

func TestOpenSignedURL_StatusHandling(t *testing.T) {
	t.Run("2xx returns the body unread", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("archive-bytes"))
		}))
		defer srv.Close()

		body, err := openSignedURL(context.Background(), srv.URL)
		if err != nil {
			t.Fatalf("openSignedURL: %v", err)
		}
		defer body.Close() //nolint:errcheck // test cleanup
		got, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(got) != "archive-bytes" {
			t.Errorf("body = %q, want %q", got, "archive-bytes")
		}
	})

	t.Run("non-2xx is an error carrying the body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("SignatureDoesNotMatch"))
		}))
		defer srv.Close()

		if _, err := openSignedURL(context.Background(), srv.URL); err == nil {
			t.Fatal("openSignedURL returned nil for a 403, want error")
		} else if !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
			t.Errorf("error = %q, want it to carry the response body", err)
		}
	})
}

type archiveEntry struct {
	typeflag byte
	linkname string
	body     string
}

// archiveEntries archives sourcePath and returns every tar entry by name. It
// reads the returned file directly, so it also asserts that createTarGzFile
// hands back a rewound handle whose reported size matches the file on disk.
func archiveEntries(t *testing.T, sourcePath string) map[string]archiveEntry {
	t.Helper()

	f, size, err := createTarGzFile(sourcePath)
	if err != nil {
		t.Fatalf("createTarGzFile(%q): %v", sourcePath, err)
	}
	t.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	})

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("stat archive: %v", err)
	}
	if size != fi.Size() {
		t.Errorf("reported size %d, file is %d bytes", size, fi.Size())
	}
	if size <= 0 {
		t.Fatalf("archive size = %d, want > 0", size)
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("open gzip: %v", err)
	}
	defer gr.Close() //nolint:errcheck // reading in a test

	entries := map[string]archiveEntry{}
	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read tar: %v", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read entry %q: %v", header.Name, err)
		}
		entries[header.Name] = archiveEntry{
			typeflag: header.Typeflag,
			linkname: header.Linkname,
			body:     string(body),
		}
	}
	return entries
}

func tarGzWithEntry(t *testing.T, name, body string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	header := &tar.Header{
		Name:     name,
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     int64(len(body)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func tempFileCount(t *testing.T) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "bitrise-rde-upload-*.tar.gz"))
	if err != nil {
		t.Fatalf("glob temp archives: %v", err)
	}
	return len(matches)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func seekableBody(t *testing.T, content string) (io.ReadSeeker, int64) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "body")
	writeFile(t, path, content)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open body: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f, int64(len(content))
}
