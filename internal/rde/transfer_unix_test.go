//go:build unix

package rde

import (
	"archive/tar"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// Symlinks are archived with their target resolved. Without it every symlink
// becomes a TypeSymlink entry with an empty Linkname — a link to nothing, which
// the session's tar refuses to extract — so any project tree with a symlink
// (node_modules/.bin, Pods, vendor, macOS framework trees) uploads broken.
func TestCreateTarGzFile_PreservesSymlinks(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "real.txt"), "real")
	if err := os.Symlink("real.txt", filepath.Join(src, "link.txt")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	entries := archiveEntries(t, src)

	link, ok := entries["link.txt"]
	if !ok {
		t.Fatalf("link.txt missing; entries = %v", entries)
	}
	if link.typeflag != tar.TypeSymlink {
		t.Errorf("link.txt typeflag = %d, want TypeSymlink (%d)", link.typeflag, tar.TypeSymlink)
	}
	if link.linkname != "real.txt" {
		t.Errorf("link.txt linkname = %q, want %q", link.linkname, "real.txt")
	}
	if _, ok := entries["real.txt"]; !ok {
		t.Errorf("real.txt missing; entries = %v", entries)
	}
}

// tar.FileInfoHeader rejects a socket outright, which would abort the whole
// upload over one stray file, so unsupported types are skipped instead.
func TestCreateTarGzFile_SkipsUnsupportedFileTypes(t *testing.T) {
	// Not t.TempDir(): its path embeds the test name, which overruns the ~104
	// byte sun_path limit for a unix socket on darwin.
	src, err := os.MkdirTemp("", "rde")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(src) })

	writeFile(t, filepath.Join(src, "keep.txt"), "keep")

	ln, err := net.Listen("unix", filepath.Join(src, "sock"))
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	defer ln.Close() //nolint:errcheck // test cleanup

	if err := syscall.Mkfifo(filepath.Join(src, "fifo"), 0o600); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	entries := archiveEntries(t, src)

	if got, ok := entries["keep.txt"]; !ok || got.body != "keep" {
		t.Errorf("keep.txt = %+v (present=%v), want body %q", got, ok, "keep")
	}
	for _, skipped := range []string{"sock", "fifo"} {
		if _, ok := entries[skipped]; ok {
			t.Errorf("%s was archived, want it skipped", skipped)
		}
	}
}

// A symlinked directory as the upload source used to archive the link itself:
// os.Stat follows it so IsDir() is true, but filepath.WalkDir lstats its root
// and never descends, emitting one entry named "." with the local target as its
// Linkname and none of the tree.
func TestCreateTarGzFile_SymlinkedDirectorySource(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "realdir")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(real, "inside.txt"), "inside")

	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	entries := archiveEntries(t, link)

	if got, ok := entries["inside.txt"]; !ok || got.body != "inside" {
		t.Errorf("inside.txt = %+v (present=%v), want body %q; entries = %v", got, ok, "inside", entries)
	}
	if got, ok := entries["."]; ok && got.typeflag == tar.TypeSymlink {
		t.Errorf(`entry "." is a symlink (linkname %q); the source link should have been resolved`, got.linkname)
	}
}
