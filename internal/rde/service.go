// Package rde is the business-logic layer over internal/rdeapi: CLI-stable
// snake_case DTOs, validation, and wait loops. The fromAPI mappers convert
// wire-format DTOs from internal/rdeapi — they're the only place where
// backend renames affect --output json.
package rde

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

// IsNotFound reports whether err is an RDE API 404 — the resource was deleted
// or never existed. Callers use it to distinguish a gone session from a
// transient failure.
func IsNotFound(err error) bool {
	var apiErr *rdeapi.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Service exposes RDE operations to the cli command layer.
type Service struct {
	client *rdeapi.Client
}

// NewService returns a Service backed by the given RDE client. The client
// must be non-nil — every method makes a network call.
func NewService(client *rdeapi.Client) *Service {
	return &Service{client: client}
}

// parseTime is the shared timestamp parser. Backend emits RFC3339 strings;
// empty input round-trips as a nil pointer so JSON output omits the field.
func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	u := t.UTC()
	return &u
}

// errClient is the canned "client not configured" error every method
// guards against. Kept here to avoid copy-paste across files.
func errClient() error { return fmt.Errorf("RDE client not configured") }

// firstNonEmpty returns a if it is non-empty, otherwise b. Used by the
// fromAPI mappers to prefer the stack id and fall back to a legacy image id
// for records snapshotted before the stack id was populated.
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// statusFromAPI converts the backend's SESSION_STATUS_* enum into a stable
// lowercase string. Falls back to the raw value (lowercased, prefix-stripped)
// for any status added after this code was written, so new statuses don't
// break callers — they just see a value that's still recognizable.
func statusFromAPI(s string) string {
	if s == "" {
		return ""
	}
	const prefix = "SESSION_STATUS_"
	v := strings.TrimPrefix(s, prefix)
	if v == "UNSPECIFIED" {
		return ""
	}
	return strings.ToLower(v)
}

// agentStatusFromAPI strips the AGENT_SESSION_STATUS_ prefix similarly.
func agentStatusFromAPI(s string) string {
	if s == "" {
		return ""
	}
	const prefix = "AGENT_SESSION_STATUS_"
	v := strings.TrimPrefix(s, prefix)
	if v == "UNSPECIFIED" {
		return ""
	}
	return strings.ToLower(v)
}

// Persistent-disk status values as produced by diskStatusFromAPI (the
// PERSISTENT_DISK_STATUS_ prefix stripped and lowercased). Only terminated
// sessions carry a disk status; running/active sessions report "". The disk
// is what a terminated session is restored from, so its status determines
// whether `rde session restore` can succeed.
const (
	DiskStatusAvailable       = "available"
	DiskStatusUnavailableSoon = "unavailable_soon"
	DiskStatusUnavailable     = "unavailable"
)

// diskStatusFromAPI strips the PERSISTENT_DISK_STATUS_ prefix.
func diskStatusFromAPI(s string) string {
	if s == "" {
		return ""
	}
	const prefix = "PERSISTENT_DISK_STATUS_"
	v := strings.TrimPrefix(s, prefix)
	if v == "UNSPECIFIED" {
		return ""
	}
	return strings.ToLower(v)
}

// looksLikeUUID is a syntactic check (8-4-4-4-12 hex). Cheap and good
// enough to distinguish IDs from names — the server validates the real
// format on its side.
func looksLikeUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
			if !isHex {
				return false
			}
		}
	}
	return true
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
