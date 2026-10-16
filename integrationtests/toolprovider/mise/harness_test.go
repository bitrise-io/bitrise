//go:build linux_and_mac
// +build linux_and_mac

package mise

import "testing"

// isolatedCache gives a mise provider its own cache dir. The tests run in parallel, and with
// the one cache dir they shared, `mise latest --installed` came back empty in CI.
func isolatedCache(t *testing.T) map[string]string {
	return map[string]string{"MISE_CACHE_DIR": t.TempDir()}
}
