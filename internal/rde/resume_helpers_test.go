package rde

import (
	"errors"
	"net/http"
	"testing"

	"github.com/bitrise-io/bitrise/v2/internal/rdeapi"
)

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(&rdeapi.APIError{StatusCode: http.StatusNotFound}) {
		t.Error("404 APIError should be not-found")
	}
	if IsNotFound(&rdeapi.APIError{StatusCode: http.StatusBadGateway}) {
		t.Error("502 APIError should not be not-found")
	}
	if IsNotFound(errors.New("network down")) {
		t.Error("plain error should not be not-found")
	}
	if IsNotFound(nil) {
		t.Error("nil should not be not-found")
	}
}

func TestSessionResumable(t *testing.T) {
	for _, tc := range []struct {
		name string
		sess Session
		want bool
	}{
		{"running", Session{Status: "running"}, true},
		{"terminated with disk", Session{Status: "terminated", PersistentDiskStatus: DiskStatusAvailable}, true},
		{"terminated disk soon-gone", Session{Status: "terminated", PersistentDiskStatus: DiskStatusUnavailableSoon}, true},
		{"terminated disk gone", Session{Status: "terminated", PersistentDiskStatus: DiskStatusUnavailable}, false},
		{"failed disk gone", Session{Status: "failed", PersistentDiskStatus: DiskStatusUnavailable}, false},
		{"starting", Session{Status: "starting"}, true},
	} {
		if got := tc.sess.Resumable(); got != tc.want {
			t.Errorf("%s: Resumable() = %v, want %v", tc.name, got, tc.want)
		}
	}
}
