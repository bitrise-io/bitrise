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
		name             string
		repoInfoProvider RepoInfoProvider
		fileReader       FileReader
		mainConfigPth    string
		wantConfig       string
		wantErr          string
	}{
		{
			name: "Circular dependency is not allowed",
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
			},
			fileReader: mockFileReader{
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
			wantErr:       "circular includes detected: repo:https://github.com/bitrise-io/example.git,bitrise.yml@commit:016883ca9498f75d03cd45c0fa400ad9f8141edf -> repo:https://github.com/bitrise-io/example.git,module_1.yml@commit:016883ca9498f75d03cd45c0fa400ad9f8141edf -> repo:https://github.com/bitrise-io/example.git,module_2.yml@commit:016883ca9498f75d03cd45c0fa400ad9f8141edf -> repo:https://github.com/bitrise-io/example.git,module_1.yml@commit:016883ca9498f75d03cd45c0fa400ad9f8141edf",
		},
		{
			name: "Max 10 include items are allowed",
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
				err: nil,
			},
			fileReader: mockFileReader{
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
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
				err: nil,
			},
			fileReader: mockFileReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(fmt.Sprintf(`format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
%s`, strings.Repeat("- path: path_1.yml\n", 10))),
					"path_1.yml": []byte(`include:
- path: path_2.yml`),
				},
			},
			mainConfigPth: "bitrise.yml",
			wantErr:       "max file count (20) exceeded",
		},
		{
			name: "Max include depth is 5",
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
				err: nil,
			},
			fileReader: mockFileReader{
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
				repoInfoProvider: tt.repoInfoProvider,
				fileReader:       tt.fileReader,
				logger:           logV2.NewLogger(),
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
		name             string
		repoInfoProvider RepoInfoProvider
		fileReader       FileReader
		mainConfigPth    string
		wantConfig       string
		wantErr          string
	}{
		{
			name: "Merges local config module",
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
				err: nil,
			},
			fileReader: mockFileReader{
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
			name: "Merges remote config module",
			repoInfoProvider: mockRepoInfoProvider{
				repoInfo: &RepoInfo{
					DefaultRemoteURL: "https://github.com/bitrise-io/example.git",
					Branch:           "main",
					Commit:           "016883ca9498f75d03cd45c0fa400ad9f8141edf",
				},
				err: nil,
			},
			fileReader: mockFileReader{
				fileSystemFiles: map[string][]byte{
					"bitrise.yml": []byte(`
format_version: "15"
default_step_lib_source: https://github.com/bitrise-io/bitrise-steplib.git

include:
- path: containers.yml
  repository: https://github.com/bitrise-io/examples-yamls.git
  branch: dev`),
				},
				repoFilesOnBranch: map[string]map[string]map[string][]byte{
					"https://github.com/bitrise-io/examples-yamls.git": {
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
				repoInfoProvider: tt.repoInfoProvider,
				fileReader:       tt.fileReader,
				logger:           logV2.NewLogger(),
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

type mockRepoInfoProvider struct {
	repoInfo *RepoInfo
	err      error
}

func (m mockRepoInfoProvider) GetRepoInfo(repoPth string) (*RepoInfo, error) {
	return m.repoInfo, m.err
}

type mockFileReader struct {
	fileSystemFiles   map[string][]byte
	fileSystemErr     error
	repoFilesOnCommit map[string]map[string]map[string][]byte
	repoFilesOnTag    map[string]map[string]map[string][]byte
	repoFilesOnBranch map[string]map[string]map[string][]byte
	repoErr           error
}

func (m mockFileReader) ReadFileFromFileSystem(name string) ([]byte, error) {
	return m.fileSystemFiles[name], m.fileSystemErr
}

func (m mockFileReader) ReadFileFromGitRepository(repository string, branch string, commit string, tag string, path string) ([]byte, error) {
	if commit != "" {
		filesInRepo, ok := m.repoFilesOnCommit[repository]
		if !ok {
			return nil, m.repoErr
		}
		filesOnCommit, ok := filesInRepo[commit]
		if !ok {
			return nil, m.repoErr
		}
		return filesOnCommit[path], m.repoErr
	} else if tag != "" {
		filesInRepo, ok := m.repoFilesOnTag[repository]
		if !ok {
			return nil, m.repoErr
		}
		filesOnTag, ok := filesInRepo[tag]
		if !ok {
			return nil, m.repoErr
		}
		return filesOnTag[path], m.repoErr
	}
	filesInRepo, ok := m.repoFilesOnBranch[repository]
	if !ok {
		return nil, m.repoErr
	}
	filesOnBranch, ok := filesInRepo[branch]
	if !ok {
		return nil, m.repoErr
	}
	return filesOnBranch[path], m.repoErr
}
