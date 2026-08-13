package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// DefaultStackID is the stack used by Create when opts.StackID is empty. Any
// valid stack ID for the org will do — this one is broadly available — and
// callers can override it via --stack.
const DefaultStackID = "ubuntu-resolute-26.04-bitrise-2026-android"

// DefaultProjectType is the project_type sent to /finish when opts.ProjectType
// is empty. "other" is a safe minimal preset.
const DefaultProjectType = "other"

// DefaultProvider is the provider sent to /apps/register when opts.Provider is
// "" or "auto". The website's "Add new app" flow sends "custom" regardless of
// host, and we mirror that so app creation doesn't trip over GitHub-App
// ownership checks; override with --provider github (etc.) when the repo is
// linked to Bitrise via that provider's app integration.
const DefaultProvider = "custom"

// FlowTypeCLI is the analytics attribution sent on register/finish so the
// server can distinguish CLI-driven app creation from website-driven flows.
const FlowTypeCLI = "cli"

// DefaultBranchFallback is the branch name sent to /apps/register when none
// can be detected from the local git checkout.
const DefaultBranchFallback = "main"

// CreateOptions are the inputs for Service.Create. Empty fields trigger
// auto-detection (git, single-workspace pick) where applicable; RepoURL
// produces an error if it can't be filled in.
type CreateOptions struct {
	RepoURL     string
	Branch      string
	Title       string
	Provider    string // "auto" or one of the API's accepted values
	OrgSlug     string // empty → fall back to single-workspace auto-detect
	StackID     string
	ProjectType string
	Public      bool

	// BitriseYML is the YAML to upload after Finish. Empty means skip the
	// upload and let the server preset chosen by ProjectType take effect.
	BitriseYML string
}

// CreateResult is what Service.Create returns once the app is registered and
// finished. BitriseYMLUploaded is true when the upload step ran.
type CreateResult struct {
	Slug              string `json:"id" yaml:"id"`
	Title             string `json:"title" yaml:"title"`
	RepoURL           string `json:"repo_url" yaml:"repo_url"`
	DefaultBranch     string `json:"default_branch" yaml:"default_branch"`
	BuildTriggerToken string `json:"build_trigger_token" yaml:"build_trigger_token"`

	OrgSlug            string `json:"workspace_id" yaml:"workspace_id"`
	StackID            string `json:"stack_id" yaml:"stack_id"`
	ProjectType        string `json:"project_type" yaml:"project_type"`
	BitriseYMLUploaded bool   `json:"bitrise_yml_uploaded" yaml:"bitrise_yml_uploaded"`
}

// Create runs the register → finish → (optional) upload sequence on the
// Bitrise API and returns the new app's identifying details.
//
// Required: opts.RepoURL (or a detectable git remote) and a workspace (via
// opts.OrgSlug or single-workspace detection). Defaults are applied for
// Branch, Provider, StackID, ProjectType.
func (s *Service) Create(ctx context.Context, opts CreateOptions) (CreateResult, error) {
	if opts.RepoURL == "" {
		return CreateResult{}, errors.New("repo URL is required (pass --repo-url or run inside a git repo with an 'origin' remote)")
	}
	provider, err := resolveProvider(opts.Provider)
	if err != nil {
		return CreateResult{}, err
	}

	branch := opts.Branch
	if branch == "" {
		branch = DefaultBranchFallback
	}
	title := opts.Title
	if title == "" {
		title = deriveTitle(opts.RepoURL)
	}
	stackID := opts.StackID
	if stackID == "" {
		stackID = DefaultStackID
	}
	projectType := opts.ProjectType
	if projectType == "" {
		projectType = DefaultProjectType
	}

	orgSlug := opts.OrgSlug
	if orgSlug == "" {
		orgSlug, err = s.autoDetectOrg(ctx)
		if err != nil {
			return CreateResult{}, err
		}
	}

	reg, err := s.client.RegisterApp(ctx, bitriseapi.RegisterAppRequest{
		RepoURL:           opts.RepoURL,
		OrganizationSlug:  orgSlug,
		Provider:          provider,
		IsPublic:          opts.Public,
		Title:             title,
		DefaultBranchName: branch,
		FlowType:          FlowTypeCLI,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("register app: %w", err)
	}

	fin, err := s.client.FinishApp(ctx, reg.Slug, bitriseapi.FinishAppRequest{
		StackID:     stackID,
		Mode:        "manual",
		ProjectType: projectType,
		Config:      configIDForProjectType(projectType),
		FlowType:    FlowTypeCLI,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("finish app %s: %w", reg.Slug, err)
	}

	res := CreateResult{
		Slug:              reg.Slug,
		Title:             title,
		RepoURL:           opts.RepoURL,
		DefaultBranch:     fin.BranchName,
		BuildTriggerToken: fin.BuildTriggerToken,
		OrgSlug:           orgSlug,
		StackID:           stackID,
		ProjectType:       projectType,
	}
	if res.DefaultBranch == "" {
		res.DefaultBranch = branch
	}

	if opts.BitriseYML != "" {
		// UpdateAppBitriseYML expects the YAML already parsed into a Go
		// value, matching internal/yml.Service.Update's own handling of the
		// same endpoint — reused here rather than adding a second client
		// method for the same POST /apps/{slug}/bitrise.yml call.
		var parsed any
		if err := yaml.Unmarshal([]byte(opts.BitriseYML), &parsed); err != nil {
			return res, fmt.Errorf("parse bitrise.yml: %w", err)
		}
		if err := s.client.UpdateAppBitriseYML(ctx, reg.Slug, parsed); err != nil {
			return res, fmt.Errorf("upload bitrise.yml: %w", err)
		}
		res.BitriseYMLUploaded = true
	}

	return res, nil
}

// autoDetectOrg fetches the user's workspaces and returns the slug when
// there's exactly one. 0 or 2+ workspaces produce a friendly error.
func (s *Service) autoDetectOrg(ctx context.Context) (string, error) {
	orgs, err := s.client.Organizations(ctx)
	if err != nil {
		return "", fmt.Errorf("list workspaces: %w", err)
	}
	switch len(orgs) {
	case 0:
		return "", errors.New("no workspaces found for this account — create one in the Bitrise dashboard, or pass --workspace")
	case 1:
		return orgs[0].Slug, nil
	default:
		names := make([]string, 0, len(orgs))
		for _, o := range orgs {
			names = append(names, fmt.Sprintf("  %s (%s)", o.Slug, o.Name))
		}
		return "", fmt.Errorf("multiple workspaces available — pass --workspace. Available:\n%s", strings.Join(names, "\n"))
	}
}

// resolveProvider validates an explicit --provider value, or returns
// DefaultProvider when "auto" or "" is given.
func resolveProvider(explicit string) (string, error) {
	switch explicit {
	case "", "auto":
		return DefaultProvider, nil
	case "github", "gitlab", "bitbucket", "custom":
		return explicit, nil
	default:
		return "", fmt.Errorf("unknown provider %q (valid: auto, github, gitlab, bitbucket, custom)", explicit)
	}
}

// configIDForProjectType returns the preset config_id the server expects for
// a given project_type. The values mirror the server's own preset list, so
// they must stay in sync with it. Unknown project_types fall through to
// "other-config", the same fallback the server applies when the field is
// omitted.
func configIDForProjectType(projectType string) string {
	switch projectType {
	case "android":
		return "default-android-config"
	case "cordova":
		return "default-cordova-config"
	case "fastlane":
		return "default-fastlane-ios-config"
	case "flutter":
		return "flutter-config-test-ios-android-web-0"
	case "ionic":
		return "default-ionic-config"
	case "ios":
		return "default-ios-config"
	case "java":
		return "default-java-gradle-config"
	case "kotlin-multiplatform":
		return "default-kotlin-multiplatform-config"
	case "macos":
		return "default-macos-config"
	case "node-js":
		return "default-node-js-npm-config"
	case "python":
		return "default-python-pip-config"
	case "react-native":
		return "default-react-native-config"
	case "ruby":
		return "default-ruby-config"
	default:
		return "other-config"
	}
}

// deriveTitle pulls a human-readable title from a repo URL: the last path
// segment with any ".git" suffix removed. Returns "" when nothing parseable
// remains.
func deriveTitle(repoURL string) string {
	clean := repoURL
	// Strip any query/fragment before the ".git" suffix, so a URL like
	// "https://host/repo.git?ref=x" still yields "repo".
	if i := strings.IndexAny(clean, "?#"); i >= 0 {
		clean = clean[:i]
	}
	clean = strings.TrimSuffix(clean, ".git")
	if u, err := url.Parse(clean); err == nil && u.Path != "" {
		base := path.Base(u.Path)
		if base != "." && base != "/" {
			return base
		}
	}
	if i := strings.LastIndexAny(clean, ":/"); i >= 0 && i+1 < len(clean) {
		return clean[i+1:]
	}
	return ""
}

// GitDetector detects the cwd's git remote URL and current branch. The
// default implementation shells out to `git`. Tests inject a stub.
type GitDetector interface {
	RemoteURL(ctx context.Context) (string, error)
	CurrentBranch(ctx context.Context) (string, error)
}

// ExecGitDetector is the default GitDetector. It runs the `git` CLI in the
// current working directory.
type ExecGitDetector struct{}

// RemoteURL returns the URL of the "origin" remote, or "" if not in a git
// repo / no origin is set. Errors from git itself are returned as-is.
func (ExecGitDetector) RemoteURL(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "remote", "get-url", "origin").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentBranch returns the name of the currently checked-out branch, or ""
// when detached or not in a git repo.
func (ExecGitDetector) CurrentBranch(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
