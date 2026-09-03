package workarounds

import (
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/hashicorp/go-version"
)

// First mise release where precompiled Ruby is the default. Older releases support it, but
// only behind the global experimental flag.
const rubyPrecompiledDefaultMiseVersion = "2026.8.0"

// GetRubyPrecompiledEnv returns the env var that lets mise install Ruby from a precompiled
// binary instead of building it with ruby-build. Returns empty strings when not needed.
//
// This matters more than install time: on macOS 26.6 with Xcode 27 a ruby-build source build
// segfaults, so the fallback is a dead path there, not just a slow one.
//
// Not ruby.compile=false, which also ungates precompiled binaries but turns a missing binary
// into a hard failure instead of a source build. jdx/ruby publishes no 3.1.x, so that would
// break every 3.1.x install.
//
// Without this the precompiled path is reached only as a side effect of canBeInstalledWithNix()
// persisting experimental=true for the nixpkgs backend, which needs macOS plus nix and so never
// happens on Linux. See https://mise.jdx.dev/lang/ruby.html
func GetRubyPrecompiledEnv(toolName provider.ToolID, miseVersion string, silent bool) (string, string) {
	if !ShouldEnableRubyPrecompiled(toolName, miseVersion) {
		return "", ""
	}

	if !silent {
		log.Debugf("[TOOLPROVIDER] Enabling mise experimental mode to allow precompiled Ruby binaries")
	}

	return "MISE_EXPERIMENTAL", "true"
}

// ShouldEnableRubyPrecompiled reports whether MISE_EXPERIMENTAL is needed for a Ruby install.
// Mise versions predating the feature entirely ignore the flag, so the upper bound is the only
// condition worth checking.
func ShouldEnableRubyPrecompiled(toolName provider.ToolID, miseVersion string) bool {
	// Matches "ruby" and the backend qualified "nixpkgs:ruby".
	if !strings.HasSuffix(string(toolName), "ruby") {
		return false
	}

	miseVer, err := version.NewVersion(strings.TrimPrefix(miseVersion, "v"))
	if err != nil {
		return false
	}

	return miseVer.LessThan(version.Must(version.NewVersion(rubyPrecompiledDefaultMiseVersion)))
}
