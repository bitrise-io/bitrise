package env

import (
	"errors"
	"fmt"
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

// Repository abstracts read/write access to process environment variables.
// Implementations should be safe to replace in tests (e.g. with an
// in-memory fake) without touching the real process environment.
type Repository interface {
	// List returns the current environment as "KEY=VALUE" entries, like
	// os.Environ().
	List() []string
	// Unset removes key from the environment.
	Unset(key string) error
	// Get returns the value of key, or "" when unset.
	Get(key string) string
	// Set assigns value to key.
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

// Getter is the minimal interface required to read an environment variable.
// Callers should prefer defining their own local interface at the use site;
// this type exists so the helpers in this package can stay small.
type Getter interface {
	Get(key string) string
}

// GetSetter is the minimal interface required to read and write environment
// variables. Callers should prefer defining their own local interface at the
// use site; this type exists so the helpers in this package can stay small.
type GetSetter interface {
	Get(key string) string
	Set(key, value string) error
}

// GetOrDefault returns r.Get(key) when non-empty, else def.
func GetOrDefault(r Getter, key, def string) string {
	if v := r.Get(key); v != "" {
		return v
	}
	return def
}

// Required returns r.Get(key) or an error when it is unset or empty.
func Required(r Getter, key string) (string, error) {
	if v := r.Get(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("required environment variable (%s) not provided", key)
}

// FlagOrEnv returns *flag when it points to a non-empty string, else r.Get(key).
// A nil pointer is treated as unset.
func FlagOrEnv(r Getter, flag *string, key string) string {
	if flag != nil && *flag != "" {
		return *flag
	}
	return r.Get(key)
}

// Revokable sets key to value on r and returns a function that restores the
// previous value when invoked.
func Revokable(r GetSetter, key, value string) (func() error, error) {
	orig := r.Get(key)
	revoke := func() error { return r.Set(key, orig) }
	return revoke, r.Set(key, value)
}

// RevokableMany sets every key in envs on r and returns a revoke function that
// restores the previous values. If any Set fails, every key already written is
// restored before returning on a best-effort basis; the returned error wraps
// both the Set failure and any restore failures, and the returned revoke is
// a no-op.
func RevokableMany(r GetSetter, envs map[string]string) (func() error, error) {
	originals := make(map[string]string, len(envs))
	revoke := func() error {
		var errs []error
		for k, v := range originals {
			if err := r.Set(k, v); err != nil {
				errs = append(errs, fmt.Errorf("restore %q: %w", k, err))
			}
		}
		return errors.Join(errs...)
	}

	for k, v := range envs {
		originals[k] = r.Get(k)
		if err := r.Set(k, v); err != nil {
			if rerr := revoke(); rerr != nil {
				return func() error { return nil }, fmt.Errorf("set %q: %w (restore failed: %w)", k, err, rerr)
			}
			return func() error { return nil }, fmt.Errorf("set %q: %w", k, err)
		}
	}
	return revoke, nil
}

// Scoped sets key to value on r, invokes fn, then restores the previous value.
// The restore runs even if fn panics, and even when the initial Set reports
// an error (a GetSetter may leave state written and still return an error).
func Scoped(r GetSetter, key, value string, fn func()) (err error) {
	revoke, setErr := Revokable(r, key, value)
	defer func() {
		if rerr := revoke(); err == nil {
			err = rerr
		}
	}()
	if setErr != nil {
		return setErr
	}
	fn()
	return nil
}

// ScopedMany applies every entry in envs on r, invokes fn, then restores all
// previous values. Restore runs even if fn panics. On setup failure
// RevokableMany has already restored every key it touched and returns a
// no-op revoke, so the deferred call is a safe cleanup either way.
func ScopedMany(r GetSetter, envs map[string]string, fn func()) (err error) {
	revoke, setErr := RevokableMany(r, envs)
	defer func() {
		if rerr := revoke(); err == nil {
			err = rerr
		}
	}()
	if setErr != nil {
		return setErr
	}
	fn()
	return nil
}
