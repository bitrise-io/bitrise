package steplib

import (
	"fmt"
	"testing"
)


func TestValidateHash(t *testing.T) {
	tests := []struct {
		name string
		filePath     string
		expectedHash string
		expectedErr  error
	}{
		{
			name: "Valid hash",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2",
			expectedErr:  nil,
		},
		{
			name: "Hash mismatch",
			filePath:     "testdata/file.txt",
			expectedHash: "sha256-1234567890abcdef",
			expectedErr:  fmt.Errorf("hash mismatch: expected sha256-1234567890abcdef, got sha256-f2040af3939f5033be8ca9b363055b3e53107c4688ba39b71d4529869a9cc9b2"),
		},
		{
			name: "Nonexistent file",
			filePath:     "testdata/nonexistent.txt",
			expectedHash: "sha256-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("open testdata/nonexistent.txt: no such file or directory"),
		},
		{
			name: "Empty hash",
			filePath:     "testdata/file.txt",
			expectedHash: "",
			expectedErr:  fmt.Errorf("hash is empty"),
		},
		{
			name: "Invalid hash type",
			filePath:     "testdata/file.txt",
			expectedHash: "md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f",
			expectedErr:  fmt.Errorf("only SHA256 hashes supported at this time, make sure to prefix the hash with `sha256-`. Found hash value: md5-3b6b4f1e2e8b8a9e4f7a4b5e6c7d8e9f"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHash(tt.filePath, tt.expectedHash)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("unexpected error: %s", err)
			} else if err == nil && tt.expectedErr != nil {
				t.Errorf("expected error: %s, but got nil", tt.expectedErr)
			} else if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("expected error: %s, but got: %s", tt.expectedErr, err)
			}
		})
	}
}
