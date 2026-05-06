package toolprovider

import (
	"slices"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestSupportedTools_MiseCoreToolConsistency(t *testing.T) {
	supported := SupportedTools()

	// Build a set of mise core tools resolved to canonical names.
	miseCoreCanonical := map[string]bool{}
	for _, tool := range mise.CoreTools() {
		canonical := string(alias.GetCanonicalToolID(provider.ToolID(tool)))
		miseCoreCanonical[canonical] = true
	}

	t.Run("every mise core tool is in SupportedTools", func(t *testing.T) {
		for tool := range miseCoreCanonical {
			assert.Contains(t, supported, tool,
				"mise core tool %q (canonical) is missing from SupportedTools — add it or document why it was excluded", tool)
		}
	})

	t.Run("every SupportedTools entry is a mise core tool or an acknowledged exception", func(t *testing.T) {
		for _, tool := range supported {
			if miseCoreCanonical[tool] {
				continue
			}
			assert.Contains(t, nonMiseCoreExceptions, tool,
				"SupportedTools entry %q is not a mise core tool and not in nonMiseCoreExceptions — add it to the exceptions list or remove it", tool)
		}
	})

	t.Run("every exception is actually in SupportedTools", func(t *testing.T) {
		for _, tool := range nonMiseCoreExceptions {
			assert.True(t, slices.Contains(supported, tool),
				"nonMiseCoreExceptions entry %q is not in SupportedTools — remove it from exceptions", tool)
		}
	})
}
