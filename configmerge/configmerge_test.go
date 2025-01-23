package configmerge

import (
	"fmt"
	"strings"
	"testing"

	logV2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestMerger_MergeConfig_Validation(t *testing.T) {
	tests := []struct {
		name          string
		configReader  ConfigReader
		mainConfigPth string
		wantConfig    string
		wantErr       string
	}{
		{
			name: "Max file size is 1MB",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(strings.Repeat(" ", MaxFileSizeBytes+1)),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "max file size (1048576 bytes) exceeded in file bitrise.yml",
		},
		{
			name: "Circular dependency is not allowed",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: module_1.yml`),
					"module_1.yml": []byte(`include:
- path: module_2.yml`),
					"module_2.yml": []byte(`include:
- path: module_1.yml`),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "circular reference detected: bitrise.yml -> module_1.yml -> module_2.yml -> module_1.yml",
		},
		{
			name: "Max 10 include items are allowed",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(fmt.Sprintf(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
%s`, strings.Repeat("- path: path_1.yml\n", MaxIncludeCountPerFile+1))),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "max include count (10) exceeded",
		},
		{
			name: "Max 20 config files are allowed",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(fmt.Sprintf(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
%s`, strings.Repeat("- path: path_1.yml\n", 10))),
					"path_1.yml": []byte(`include:
- path: path_2.yml`),
					"path_2.yml": []byte(``),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "max file count (20) exceeded",
		},
		{
			name: "Max include depth is 5",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: module_1.yml
  repository: http://github.com/bitrise-io/bitrise-yamls.git
  branch: main`),
				},
				repoFilesOnBranch: map[string]map[string]map[string][]byte{
					"http://github.com/bitrise-io/bitrise-yamls.git": {
						"main": {
							"module_1.yml": []byte(`include:
- path: module_2.yml
  repository: http://github.com/bitrise-io/bitrise-yamls.git
  branch: main`),
							"module_2.yml": []byte(`include:
- path: module_3.yml
  repository: http://github.com/bitrise-io/bitrise-yamls.git
  branch: main`),
							"module_3.yml": []byte(`include:
- path: module_4.yml
  repository: http://github.com/bitrise-io/bitrise-yamls.git
  branch: main`),
							"module_4.yml": []byte(`include:
- path: module_5.yml
  repository: http://github.com/bitrise-io/bitrise-yamls.git
  branch: main`),
							"module_5.yml": []byte(``),
						},
					},
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "max include depth (5) exceeded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				configReader: tt.configReader,
				logger:       logV2.NewLogger(),
			}
			got, _, err := m.MergeConfig(tt.mainConfigPth)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantConfig, got, got)
		})
	}

}

func TestMerger_MergeConfig(t *testing.T) {
	tests := []struct {
		name          string
		configReader  ConfigReader
		mainConfigPth string
		wantConfig    string
		wantErr       string
	}{
		{
			name: "Merges local config module",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: containers.yml`),
					"containers.yml": []byte(`containers:
  golang:
    image: golang:1.22`),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantConfig: `containers:
  golang:
    image: golang:1.22
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
format_version: "15"
`,
		},
		{
			name: "Include path is relative to the main config file dir (repo root)",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: configs/ci/module_1.yml`),
					"configs/ci/module_1.yml": []byte(`include:
- path: configs/ci/module_2.yml`),
					"configs/ci/module_2.yml": []byte(`workflows:
  print_hello:
    steps:
    - script:
        inputs:
        - content: echo "Hello World!"`),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantConfig: `default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
format_version: "15"
workflows:
  print_hello:
    steps:
    - script:
        inputs:
        - content: echo "Hello World!"
`,
		},
		{
			name: "Merges remote config module",
			configReader: mockConfigReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`
format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: containers.yml
  repository: examples-yamls
  branch: dev`),
				},
				repoFilesOnBranch: map[string]map[string]map[string][]byte{
					"examples-yamls": {
						"dev": {
							"containers.yml": []byte(`
containers:
  golang:
    image: golang:1.22`),
						},
					},
				},
			},
			mainConfigPth: "bitrise.yml",
			wantConfig: `containers:
  golang:
    image: golang:1.22
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git
format_version: "15"
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				configReader: tt.configReader,
				logger:       logV2.NewLogger(),
			}
			got, _, err := m.MergeConfig(tt.mainConfigPth)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantConfig, got, got)
		})
	}
}

type mockConfigReader struct {
	fileSystemFiles   map[string][]byte
	repoFilesOnCommit map[string]map[string]map[string][]byte
	repoFilesOnTag    map[string]map[string]map[string][]byte
	repoFilesOnBranch map[string]map[string]map[string][]byte
}

func (m mockConfigReader) Read(ref ConfigReference) ([]byte, error) {
	if ref.IsLocalReference() {
		return m.readFileFromFileSystem(ref.Path)
	}
	return m.readFileFromGitRepository(ref.Repository, ref.Branch, ref.Commit, ref.Tag, ref.Path)

}

func (m mockConfigReader) CleanupRepoDirs() error {
	return nil
}

func (m mockConfigReader) readFileFromFileSystem(name string) ([]byte, error) {
	c, ok := m.fileSystemFiles[name]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", name)
	}
	return c, nil
}

func (m mockConfigReader) readFileFromGitRepository(repository string, branch string, commit string, tag string, path string) ([]byte, error) {
	var repoFiles map[string]map[string]map[string][]byte
	var checkout string
	if commit != "" {
		repoFiles = m.repoFilesOnCommit
		checkout = commit
	} else if tag != "" {
		repoFiles = m.repoFilesOnTag
		checkout = tag
	} else {
		repoFiles = m.repoFilesOnBranch
		checkout = branch
	}

	filesInRepo, ok := repoFiles[repository]
	if !ok {
		return nil, fmt.Errorf("repo not found: %s", repository)
	}
	filesOnCommit, ok := filesInRepo[checkout]
	if !ok {
		return nil, fmt.Errorf("checkout param not found: %s", checkout)
	}
	c, ok := filesOnCommit[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return c, nil
}
