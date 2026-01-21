package nixpkgs

import (
	"os/exec"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

const (
	PluginGitURL = "https://github.com/bitrise-io/mise-nixpkgs-plugin.git"
	PluginName   = "nixpkgs"
)

func ShouldUseBackend(request provider.ToolRequest, silent bool) bool {
	if request.ToolName != "ruby" {
		if !silent {
			log.Debugf("[TOOLPROVIDER] The mise-nixpkgs backend is only enabled for Ruby for now. Using core plugin to install %s", request.ToolName)
		}
		return false
	}

	if !isNixAvailable(silent) {
		if !silent {
			log.Debugf("[TOOLPROVIDER] Nix is not available on the system, cannot use nixpkgs backend for %s", request.ToolName)
		}
		return false
	}

	if !silent {
		log.Debugf("[TOOLPROVIDER] Nix backend is available for %s", request.ToolName)
	}
	return true
}

func isNixAvailable(silent bool) bool {
	_, err := exec.LookPath("nix")
	if err != nil {
		return false
	}

	cmd := exec.Command("nix", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !silent {
			log.Debugf("[TOOLPROVIDER] Exec nix --version failed: %v\nOutput: %s", err, string(out))
		}
		return false
	}

	return true
}
