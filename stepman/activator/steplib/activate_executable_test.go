package steplib

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedHash string
		expectedErr  error
	}{
		{
			name:         "Valid hash",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
			expectedErr:  nil,
		},
		{
			name:         "Hash mismatch",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-1234567890abcdef",
			expectedErr:  fmt.Errorf("hash mismatch: expected sha256-1234567890abcdef, got sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2"),
		},
		{
			name:         "Nonexistent file",
			filePath:     "testdata/nonexistent.txt",
			expectedHash: "sha256-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("open testdata/nonexistent.txt: no such file or directory"),
		},
		{
			name:         "Empty hash",
			filePath:     "testdata/file.txt",
			expectedHash: "",
			expectedErr:  fmt.Errorf("hash is empty"),
		},
		{
			name:         "Invalid hash type",
			filePath:     "testdata/file.txt",
			expectedHash: "md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHash(tt.filePath, tt.expectedHash)
			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		executable  models.Executable
		expectedURL string
	}{
		{
			name: "With custom base URL",
			env: map[string]string{
				precompiledStepsPrimaryStorageEnv: "https://custom.example.com/storage",
			},
			executable: models.Executable{
				StorageURI: "steps/step1.tar.gz",
			},
			expectedURL: "https://custom.example.com/storage/steps/step1.tar.gz",
		},
		{
			name: "With custom base URL with trailing slash",
			env: map[string]string{
				precompiledStepsPrimaryStorageEnv: "https://custom.example.com/storage/",
			},
			executable: models.Executable{
				StorageURI: "steps/step1.tar.gz",
			},
			expectedURL: "https://custom.example.com/storage/steps/step1.tar.gz",
		},
		{
			name: "With default base URL",
			env:  map[string]string{},
			executable: models.Executable{
				StorageURI: "steps/step2.tar.gz",
			},
			expectedURL: "https://storage.googleapis.com/bitrise-steplib-storage/steps/step2.tar.gz",
		},
		{
			name: "With leading slash in storage URI",
			env: map[string]string{
				precompiledStepsPrimaryStorageEnv: "https://custom.example.com/storage",
			},
			executable: models.Executable{
				StorageURI: "/steps/step3.tar.gz",
			},
			expectedURL: "https://custom.example.com/storage/steps/step3.tar.gz",
		},
		{
			name: "With multiple slashes in base URL",
			env: map[string]string{
				precompiledStepsPrimaryStorageEnv: "https://custom.example.com/storage///",
			},
			executable: models.Executable{
				StorageURI: "steps/step4.tar.gz",
			},
			expectedURL: "https://custom.example.com/storage/steps/step4.tar.gz",
		},
		{
			name: "Empty storage URI",
			env: map[string]string{
				precompiledStepsPrimaryStorageEnv: "https://custom.example.com/storage",
			},
			executable: models.Executable{
				StorageURI: "",
			},
			expectedURL: "https://custom.example.com/storage/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if val, exists := tt.env[precompiledStepsPrimaryStorageEnv]; exists {
				t.Setenv(precompiledStepsPrimaryStorageEnv, val)
			}
			url := downloadURL(tt.executable)
			require.Equal(t, tt.expectedURL, url)
		})
	}
}
