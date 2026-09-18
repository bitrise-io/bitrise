// Package workspace holds the rules for picking a workspace when the user
// didn't name one. It is a leaf package so both the CLI layer
// (cli/cmdutil.ResolveWorkspaceID) and a domain service (internal/app's app
// creation) can share one definition — and one error message — without
// internal/* depending on cli/*.
package workspace

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// EnvWorkspaceID supplies the workspace when --workspace isn't passed. Bitrise
// sets it in every build. It lives here rather than beside the other env-var
// constants in cli/cmdutil because Sole's guidance message names it, and
// internal/* must not import cli/*; cli/cmdutil re-exports it.
const EnvWorkspaceID = "BITRISE_WORKSPACE_ID"

// Sole returns the user's workspace when they have exactly one. With zero or
// 2+ workspaces it returns a friendly error, since no default can be picked
// unambiguously. This is the single definition of the "exactly one workspace"
// rule and its guidance message.
func Sole(orgs []bitriseapi.Organization) (bitriseapi.Organization, error) {
	switch len(orgs) {
	case 0:
		return bitriseapi.Organization{}, errors.New("no workspaces found for this account — create one in the Bitrise dashboard, or pass --workspace")
	case 1:
		return orgs[0], nil
	default:
		return bitriseapi.Organization{}, fmt.Errorf("multiple workspaces available — pass --workspace, set %s, or run 'bitrise config set %s <id>'. Available:\n%s",
			EnvWorkspaceID, config.KeyDefaultWorkspaceID, List(orgs))
	}
}

// Sort returns a copy of orgs sorted for human display: named workspaces
// first, alphabetically by name (case-insensitive), then any the API returned
// without a name, by slug. It's the single source of ordering for both the
// multiple-workspaces error list and the interactive picker, so the two stay
// in sync.
func Sort(orgs []bitriseapi.Organization) []bitriseapi.Organization {
	sorted := append([]bitriseapi.Organization(nil), orgs...)
	sort.Slice(sorted, func(i, j int) bool {
		ni, nj := sorted[i].Name, sorted[j].Name
		if (ni == "") != (nj == "") {
			return ni != "" // named workspaces first
		}
		if !strings.EqualFold(ni, nj) {
			return strings.ToLower(ni) < strings.ToLower(nj)
		}
		return sorted[i].Slug < sorted[j].Slug
	})
	return sorted
}

// List renders workspaces one per indented line as "name (slug)", sorted by
// name so a user can scan for the one they recognize and copy its slug.
// Workspaces the API returned without a name fall back to the bare slug and
// sort last.
func List(orgs []bitriseapi.Organization) string {
	sorted := Sort(orgs)
	lines := make([]string, len(sorted))
	for i, o := range sorted {
		if o.Name != "" {
			lines[i] = fmt.Sprintf("  %s (%s)", o.Name, o.Slug)
		} else {
			lines[i] = "  " + o.Slug
		}
	}
	return strings.Join(lines, "\n")
}
