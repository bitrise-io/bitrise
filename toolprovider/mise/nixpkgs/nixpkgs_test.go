package nixpkgs

import (
	"testing"
)

func TestIsNixAvailable(t *testing.T) {
	result := isNixAvailable()
	t.Logf("isNixAvailable() returned: %v", result)

	if !result {
		t.Fatal("✗ Nix is not available on this system")
	}

	t.Log("✓ Nix is available on this system")
}
