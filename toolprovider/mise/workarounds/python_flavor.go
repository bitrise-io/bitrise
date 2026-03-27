package workarounds

import (
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/hashicorp/go-version"
)

// GetPythonPrecompiledFlavorEnv returns the environment variable key and value for Python 3.14+
// with mise versions before 2026 to avoid missing lib directory errors.
// Returns empty strings if the workaround is not needed.
// https://mise.jdx.dev/lang/python.html#python.precompiled_flavor
// https://github.com/jdx/mise/releases/tag/v2026.3.10
func GetPythonPrecompiledFlavorEnv(toolName provider.ToolID, concreteVersion string, miseVersion string, silent bool) (string, string) {
	if !ShouldSetPythonPrecompiledFlavor(toolName, concreteVersion, miseVersion) {
		return "", ""
	}

	if !silent {
		log.Debugf("[TOOLPROVIDER] Setting MISE_PYTHON_PRECOMPILED_FLAVOR for Python %s", concreteVersion)
	}

	return "MISE_PYTHON_PRECOMPILED_FLAVOR", "install_only_stripped"
}

// shouldSetPythonPrecompiledFlavor determines if MISE_PYTHON_PRECOMPILED_FLAVOR should be set.
// This is needed for Python 3.14+ with mise versions before 2026 to avoid missing lib directory errors.
// The fix was implemented in mise v2026.3.10.
func ShouldSetPythonPrecompiledFlavor(toolName provider.ToolID, concreteVersion string, miseVersion string) bool {
	toolNameStr := string(toolName)
	if !strings.HasSuffix(toolNameStr, "python") {
		return false
	}

	miseVerStr := strings.TrimPrefix(miseVersion, "v")
	miseVer, err := version.NewVersion(miseVerStr)
	if err != nil {
		return false
	}

	// The fix is in mise v2026.3.10.
	miseFixVersion := version.Must(version.NewVersion("2026.3.10"))
	if miseVer.GreaterThanOrEqual(miseFixVersion) {
		return false
	}

	pythonVer, err := version.NewVersion(concreteVersion)
	if err != nil {
		return false
	}

	// Compare only major.minor versions to handle pre-releases like 3.14.0a1.
	segments := pythonVer.Segments()
	if len(segments) < 2 {
		return false
	}

	major := segments[0]
	minor := segments[1]

	// Python 3.14+ has the issue.
	return major > 3 || (major == 3 && minor >= 14)
}
