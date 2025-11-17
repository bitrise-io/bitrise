package nixpkgs

import (
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

func ShouldUseBackend(request provider.ToolRequest) bool {
	if request.ToolName != "ruby" {
		log.Debugf("[TOOLPROVIDER] The mise-nixpkgs backend is only enabled for Ruby for now. Using core plugin to install %s", request.ToolName)
		return false
	}

	value, ok := os.LookupEnv("BITRISE_TOOLSETUP_FAST_INSTALL")
	if !ok || strings.TrimSpace(value) != "true" {
		log.Debugf("[TOOLPROVIDER] Using core mise plugin for %s", request.ToolName)
		return false
	}

	if !isNixAvailable() {
		log.Debugf("[TOOLPROVIDER] Nix is not available on the system, cannot use nixpkgs backend for %s", request.ToolName)
		return false
	}

	log.Debugf("[TOOLPROVIDER] Using nixpkgs backend for %s as BITRISE_TOOLSETUP_FAST_INSTALL is set", request.ToolName)
	return true
}

func isNixAvailable() bool {
	_, err := exec.LookPath("nix")
	if err != nil {
		return false
	}

	cmd := exec.Command("nix", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("[TOOLPROVIDER] Exec nix --version: %s, output:\n%s", err, out)
		return false
	}

	return true
}
