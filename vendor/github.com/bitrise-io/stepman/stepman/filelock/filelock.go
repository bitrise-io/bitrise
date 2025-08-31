package filelock

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

// FileLock provides cross-platform file locking functionality
type FileLock struct {
	file *os.File
	path string
}

// NewFileLock creates a new file lock for the given path
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path,
	}
}

// Lock acquires an exclusive lock on the file with a 30-second timeout
func (fl *FileLock) Lock() error {
	return fl.lockWithTimeout(30 * time.Second)
}

// TryLock attempts to acquire a non-blocking exclusive lock
func (fl *FileLock) TryLock() error {
	return fl.lockWithTimeout(0)
}

// lockWithTimeout acquires a lock with the specified timeout
func (fl *FileLock) lockWithTimeout(timeout time.Duration) error {
	if fl.file != nil {
		return fmt.Errorf("lock already acquired")
	}

	// Ensure lock directory exists
	lockDir := filepath.Dir(fl.path)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open or create the lock file
	file, err := os.OpenFile(fl.path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Attempt to acquire the lock
	start := time.Now()
	for {
		err = fl.acquireLock(file)
		if err == nil {
			fl.file = file
			// Write PID to lock file for debugging
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			_ = file.Sync()
			return nil
		}

		if timeout == 0 {
			_ = file.Close()
			return fmt.Errorf("failed to acquire lock (non-blocking): %w", err)
		}

		if time.Since(start) >= timeout {
			_ = file.Close()
			return fmt.Errorf("failed to acquire lock within timeout: %w", err)
		}

		// Exponential backoff: 100ms, 200ms, 400ms, 800ms, then 1s
		var backoff time.Duration
		if elapsed := time.Since(start); elapsed < 1600*time.Millisecond {
			backoff = 100 * time.Millisecond * (1 << uint(elapsed/(100*time.Millisecond)%4))
		} else {
			backoff = time.Second
		}
		
		time.Sleep(backoff)
	}
}

// acquireLock performs the platform-specific lock acquisition
func (fl *FileLock) acquireLock(file *os.File) error {
	if runtime.GOOS == "windows" {
		return fl.lockWindows(file)
	}
	return fl.lockUnix(file)
}

// lockUnix acquires a lock on Unix-like systems (Linux, macOS, etc.)
func (fl *FileLock) lockUnix(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// lockWindows acquires a lock on Windows systems
func (fl *FileLock) lockWindows(file *os.File) error {
	// For Windows, we use a different approach since flock is not available
	// We'll use CreateFile with exclusive access instead
	return fmt.Errorf("windows file locking not implemented in this version")
}

// Unlock releases the file lock
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return fmt.Errorf("no lock to release")
	}

	var err error
	if runtime.GOOS == "windows" {
		err = fl.unlockWindows()
	} else {
		err = fl.unlockUnix()
	}

	// Close the file regardless of unlock result
	closeErr := fl.file.Close()
	fl.file = nil

	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close lock file: %w", closeErr)
	}

	return nil
}

// unlockUnix releases a lock on Unix-like systems
func (fl *FileLock) unlockUnix() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
}

// unlockWindows releases a lock on Windows systems
func (fl *FileLock) unlockWindows() error {
	return fmt.Errorf("windows file locking not implemented in this version")
}

// Close releases the lock and closes the file
func (fl *FileLock) Close() error {
	if fl.file == nil {
		return nil
	}
	return fl.Unlock()
}

// IsStale checks if a lock file is stale (older than 5 minutes with no active process)
func IsStale(lockPath string) bool {
	info, err := os.Stat(lockPath)
	if err != nil {
		return false
	}
	
	// If lock file is older than 5 minutes, consider it potentially stale
	if time.Since(info.ModTime()) < 5*time.Minute {
		return false
	}
	
	// Try to read PID from lock file
	file, err := os.Open(lockPath)
	if err != nil {
		return true
	}
	defer func() { _ = file.Close() }()
	
	var pid int
	if _, err := fmt.Fscanf(file, "%d", &pid); err != nil {
		return true
	}
	
	// Check if process is still running
	process, err := os.FindProcess(pid)
	if err != nil {
		return true
	}
	
	// On Unix, we can send signal 0 to check if process exists
	if runtime.GOOS != "windows" {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return true
		}
	}
	
	return false
}

// CleanupStale removes stale lock files
func CleanupStale(lockPath string) error {
	if IsStale(lockPath) {
		return os.Remove(lockPath)
	}
	return nil
}