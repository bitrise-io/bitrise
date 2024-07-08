package configmerge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merger{
				repoInfoProvider: tt.repoInfoProvider,
				fileReader:       tt.fileReader,
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
