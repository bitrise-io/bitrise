package nixpkgs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

const (
	PluginGitURL = "https://github.com/bitrise-io/mise-nixpkgs-plugin.git"
	PluginName   = "nixpkgs"
)

func ShouldUseBackend(request provider.ToolRequest) (bool, error) {
	if request.ToolName != "ruby" {
		log.Debugf("[TOOLPROVIDER] The mise-nixpkgs backend is only enabled for Ruby for now. Using core plugin to install %s", request.ToolName)
		return false, nil
	}

	value, ok := os.LookupEnv("BITRISE_TOOLSETUP_FAST_INSTALL")
	if !ok || strings.TrimSpace(value) != "true" {
		log.Debugf("[TOOLPROVIDER] Using core mise plugin for %s", request.ToolName)
		return false, nil
	}

	if !isNixAvailable() {
		log.Debugf("[TOOLPROVIDER] Nix is not available on the system, cannot use nixpkgs backend for %s", request.ToolName)
		return false, fmt.Errorf("nix not available on system")
	}

	log.Debugf("[TOOLPROVIDER] Using nixpkgs backend for %s as BITRISE_TOOLSETUP_FAST_INSTALL is set", request.ToolName)
	return true, nil
}

func isNixAvailable() bool {
	// Check for testing override first.
	if value, ok := os.LookupEnv("BITRISE_TEST_SKIP_NIX_CHECK"); ok && strings.TrimSpace(value) == "true" {
		return true
	}

	_, err := exec.LookPath("nix")
	if err != nil {
		return false
	}

	cmd := exec.Command("nix", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("[TOOLPROVIDER] Exec nix --version failed: %v\nOutput: %s", err, string(out))
		return false
	}

	return true
}
