package atomicwrite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically using the temp-then-rename pattern
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Create temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, fmt.Sprintf(".tmp-%s-*", base))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	
	tempPath := tempFile.Name()
	defer func() {
		// Clean up temp file on any error
		if tempFile != nil {
			_ = tempFile.Close()
		}
		_ = os.Remove(tempPath)
	}()
	
	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	
	// Sync to ensure data is written to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	
	// Set permissions before closing
	if err := tempFile.Chmod(perm); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}
	
	// Close temp file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tempFile = nil // Prevent cleanup in defer
	
	// Atomic rename to final location
	if err := os.Rename(tempPath, filename); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", filename, err)
	}
	
	return nil
}

// WriteBytesAtomic writes bytes to a file atomically with default permissions (0644)
func WriteBytesAtomic(filename string, data []byte) error {
	return WriteFileAtomic(filename, data, 0644)
}

// WriteJSONAtomic writes a JSON-encoded value to a file atomically
func WriteJSONAtomic(filename string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return WriteBytesAtomic(filename, data)
}

// WriteStringAtomic writes a string to a file atomically with default permissions (0644)
func WriteStringAtomic(filename string, content string) error {
	return WriteBytesAtomic(filename, []byte(content))
}