package mise

import (
	"slices"
	"strings"
	"unicode"

	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
)

// releasedVersionsInLine turns mise's ls-remote list into the provider's listing.
//
// mise lists oldest first and resolves a prefix to the last stable match in that
// list, so the list is reversed rather than sorted: semver puts 1.20.4 above
// 1.20.4-otp-29, and text order puts jruby-9 above jruby-10. Prereleases stay in,
// consumers skip them as mise does. The workflow editor relies on this order for
// its hints. Trailing dots are dropped from the prefix, so 20. lists the 20 line.
func releasedVersionsInLine(versions []string, toolName provider.ToolID, prefix string) []string {
	slices.Reverse(versions)

	prefix = strings.TrimRight(prefix, ".")
	if prefix == "" {
		return versions
	}

	var inLine []string
	for _, v := range versions {
		if isInVersionLine(v, prefix, toolName) {
			inLine = append(inLine, v)
		}
	}
	return inLine
}

// isInVersionLine mirrors how mise matches a prefix: 24.2 covers 24.2.0, not 24.20.0,
// and 1.19.0 covers 1.19.0-otp-28. A plus continues only a numeric line, as in
// 1.7.8+hotfix.4-stable, because after a name it joins flavours, as in truffleruby+graalvm.
// mise also tries the prefix with or without a leading v, so v22 covers 22.1.0.
// A vendor prefix such as temurin- runs straight into the version.
// mise's java plugin matches differently: a plus continues any line, as in
// temurin-21.0.5+11.0.LTS, and there are no v spellings.
func isInVersionLine(version, prefix string, toolName provider.ToolID) bool {
	isJava := toolName == "java"

	separators := ".-"
	if isJava || isNumericPrefix(prefix) {
		separators += "+"
	}

	heads := []string{prefix}
	if !isJava {
		heads = append(heads, vSpellings(prefix)...)
	}

	isVendorPrefix := strings.HasSuffix(prefix, "-")
	for _, head := range heads {
		if isVendorPrefix && strings.HasPrefix(version, head) {
			return true
		}
		if !isVendorPrefix && continuesLine(version, head, separators) {
			return true
		}
	}
	return false
}

func continuesLine(version, head, separators string) bool {
	if version == head {
		return true
	}

	rest, ok := strings.CutPrefix(version, head)
	// mise wants something after the separator, so 22. is not in the 22 line.
	if !ok || len(rest) < 2 {
		return false
	}
	return strings.ContainsRune(separators, rune(rest[0]))
}

func isNumericPrefix(prefix string) bool {
	if prefix[0] == 'v' || prefix[0] == 'V' {
		prefix = prefix[1:]
	}
	return prefix != "" && unicode.IsDigit(rune(prefix[0]))
}

func vSpellings(prefix string) []string {
	switch {
	case unicode.IsDigit(rune(prefix[0])):
		return []string{"v" + prefix}
	case (prefix[0] == 'v' || prefix[0] == 'V') && len(prefix) > 1:
		return []string{prefix[1:]}
	default:
		return nil
	}
}
