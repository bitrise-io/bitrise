package toolprovider

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/go-utils/v2/log/colorstring"
)

func printToolRequests(toolRequests []provider.ToolRequest) {
	log.Printf("")
	log.Info("Tool setup")
	log.Printf("Plan:")

	for _, toolRequest := range toolRequests {
		strategy := "" // Default is strict, we don't print anything
		switch toolRequest.ResolutionStrategy {
		case provider.ResolutionStrategyLatestInstalled:
			strategy = "(resolve to latest installed)"
		case provider.ResolutionStrategyLatestReleased:
			strategy = "(resolve to latest released)"
		}

		versionStr := toolRequest.UnparsedVersion
		if versionStr == "" {
			versionStr = "<unset version>"
		}
		log.Printf(
			"• %s %s %s",
			colorstring.Magenta(toolRequest.ToolName),
			colorstring.Cyan(versionStr),
			strategy,
		)
	}

	if len(toolRequests) > 0 {
		log.Printf("")
		log.Printf("Installing missing tools")
	}
}

func printInstallStart(toolRequest provider.ToolRequest) {
	versionStr := toolRequest.UnparsedVersion
	if versionStr == "" {
		versionStr = "<unset version>"
	}
	log.Printf(
		"• %s %s...",
		colorstring.Magenta(toolRequest.ToolName),
		colorstring.Cyan(versionStr),
	)
}

func printInstallResult(toolRequest provider.ToolRequest, result provider.ToolInstallResult) {
	var status string
	if result.IsAlreadyInstalled {
		status = "already installed"
	} else {
		status = "installed"
	}
	ver := ""
	if result.ConcreteVersion != toolRequest.UnparsedVersion {
		ver = fmt.Sprintf("(%s)", colorstring.Cyan(result.ConcreteVersion))
	}

	log.Printf("%s %s %s", colorstring.Green("✓"), status, ver)
	log.Printf("")
}

func printInstallError(err provider.ToolInstallError) {
	optionalLines := ""
	if err.Cause != "" {
		optionalLines += "  • " + colorstring.Magenta("Cause:") + " " + err.Cause
		optionalLines += "\n"
	}
	if err.Recommendation != "" {
		optionalLines += "  • " + colorstring.Magenta("Recommendation:") + " " + err.Recommendation
		optionalLines += "\n"
	}
	if err.RawOutput != "" {
		optionalLines += "  • " + colorstring.Magenta("Raw output:") + " " + err.RawOutput
		optionalLines += "\n"
	}

	log.Printf(
		"%s",
		colorstring.Redf("⨯ install %s %s", err.ToolName, err.RequestedVersion),
	)
	log.Print(optionalLines)
}
