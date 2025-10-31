package env

import (
	"os"
	"os/exec"
)

// CommandLocator ...
type CommandLocator interface {
	LookPath(file string) (string, error)
}

type commandLocator struct{}

// NewCommandLocator ...
func NewCommandLocator() CommandLocator {
	return commandLocator{}
}

// LookPath ...
func (l commandLocator) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Repository ...
type Repository interface {
	List() []string
	Unset(key string) error
	Get(key string) string
	Set(key, value string) error
}

// NewRepository ...
func NewRepository() Repository {
	return repository{}
}

type repository struct{}

// Get ...
func (d repository) Get(key string) string {
	return os.Getenv(key)
}

// Set ...
func (d repository) Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset ...
func (d repository) Unset(key string) error {
	return os.Unsetenv(key)
}

// List ...
func (d repository) List() []string {
	return os.Environ()
}
