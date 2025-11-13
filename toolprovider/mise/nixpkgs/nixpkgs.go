package nixpkgs

import (
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

const (
	PluginGitURL = "https://github.com/bitrise-io/mise-nixpkgs-plugin.git"
	PluginName   = "mise-nixpkgs-plugin"
)

// TODO: check if Nix works at all

func ShouldUseBackend(request provider.ToolRequest) bool {
	if request.ToolName != "ruby" {
		log.Debugf("[TOOLPROVIDER] The mise-nixpkgs backend is only enabled for Ruby for now. Using core plugin to install %s", request.ToolName)
		return false
	}

	value, ok := os.LookupEnv("BITRISE_TOOLSETUP_FAST_INSTALL")
	if ok && strings.TrimSpace(value) == "1" {
		log.Debugf("[TOOLPROVIDER] Using nixpkgs backend for %s as BITRISE_TOOLSETUP_FAST_INSTALL is set", request.ToolName)
		return true
	}
	log.Debugf("[TOOLPROVIDER] Using core plugin for %s", request.ToolName)
	return false
}
