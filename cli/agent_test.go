package cli

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanupDirs(t *testing.T) {
	tmpDir := t.TempDir()
	testCases := []struct {
		name      string
		dirs      []string
		prepareFn func()
		wantErr   bool
	}{
		{
			name: "Clean existing file",
			dirs: []string{filepath.Join(tmpDir, "file")},
			prepareFn: func() {
				err := ioutil.WriteFile(filepath.Join(tmpDir, "file"), []byte("file"), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "Clean existing directory",
			dirs: []string{filepath.Join(tmpDir, "dir")},
			prepareFn: func() {
				err := ioutil.WriteFile(filepath.Join(tmpDir, "dir", "file1"), []byte("file1"), 0644)
				require.NoError(t, err)
				err = ioutil.WriteFile(filepath.Join(tmpDir, "dir", "file2"), []byte("file2"), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name:      "Empty dir list",
			dirs:      []string{},
			prepareFn: func() {},
			wantErr:   false,
		},
		{
			name:      "Nonexistent dir",
			dirs:      []string{filepath.Join(tmpDir, "nonexistent")},
			prepareFn: func() {},
			wantErr:   false,
		},
		{
			name: "Env vars",
			dirs: []string{
				"$BITRISE_TEST_DEPLOY_DIR",
				"$BITRISE_DEPLOY_DIR/artifacts",
			},
			prepareFn: func() {
				dir1 := filepath.Join(tmpDir, "test_deploy_dir")
				dir2 := filepath.Join(tmpDir, "deploy_dir")

				t.Setenv("BITRISE_TEST_DEPLOY_DIR", dir1)
				t.Setenv("BITRISE_DEPLOY_DIR", dir2)

				err := ioutil.WriteFile(filepath.Join(dir1, "file1"), []byte("file1"), 0644)
				require.NoError(t, err)
				err = ioutil.WriteFile(filepath.Join(dir2, "file1"), []byte("file2"), 0644)
				require.NoError(t, err)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cleanupDirs(tc.dirs)

			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
			}

			for _, dir := range tc.dirs {
				_, err := ioutil.ReadDir(dir)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
