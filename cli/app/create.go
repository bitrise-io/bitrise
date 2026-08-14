package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/configs"
	internalapp "github.com/bitrise-io/bitrise/v2/internal/app"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// createFlags bundles app create's parsed flag values so runCreate can take
// a detector parameter (for git-detection injection in tests) without also
// threading nine separate flag values through it.
type createFlags struct {
	repoURL     string
	branch      string
	title       string
	provider    string
	workspace   string
	stackID     string
	projectType string
	public      bool
	bitriseYML  string
}

// NewCreateCommand returns the `app create` subcommand.
func NewCreateCommand() *cobra.Command {
	var flags createFlags

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new app on Bitrise",
		Long: fmt.Sprintf(`Register a new app on Bitrise.

Auto-detection from the current git repo:
  --repo-url     git remote get-url origin
  --branch       git symbolic-ref --short HEAD (else %q)
  --title        last path segment of the repo URL (".git" stripped)

Workspace, highest to lowest:
  --workspace WORKSPACE_ID
  $BITRISE_WORKSPACE_ID
  the default_workspace_id config key ('bitrise config set')
  auto-detected, when your account has exactly one workspace

So --workspace is only required if you belong to several workspaces and
haven't set a default.

bitrise.yml handling:
  --bitrise-yml PATH                upload that file as the app's config
  (no flag, ./bitrise.yml exists)   upload it
  (no flag, no file)                skip — server preset for --project-type applies

The new app's ID is saved as the default app_id in
~/.config/bitrise/cli/config.yml, so later commands (e.g. 'bitrise yml get')
target it without --app.`, internalapp.DefaultBranchFallback),
		Example: `  bitrise app create
  bitrise app create --repo-url https://github.com/me/proj --workspace acme
  bitrise app create --bitrise-yml ./ci/bitrise.yml --stack osx-xcode-16.0.x
  bitrise app create --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreate(cmd, internalapp.ExecGitDetector{}, flags)
		},
	}

	cmd.Flags().StringVar(&flags.repoURL, "repo-url", "", "git repo URL (default: 'git remote get-url origin' in cwd)")
	cmd.Flags().StringVar(&flags.branch, "branch", "", fmt.Sprintf("default branch (default: 'git symbolic-ref --short HEAD', else %q)", internalapp.DefaultBranchFallback))
	cmd.Flags().StringVar(&flags.title, "title", "", "app title (default: last path segment of repo URL)")
	cmd.Flags().StringVar(&flags.provider, "provider", "auto", "git provider: auto, github, gitlab, bitbucket, custom")
	cmd.Flags().StringVar(&flags.workspace, cmdutil.FlagWorkspace, "", "workspace ID to own the app (or set BITRISE_WORKSPACE_ID / default_workspace_id; auto-detected if you have exactly one)")
	cmd.Flags().StringVar(&flags.stackID, "stack", "", fmt.Sprintf("build stack ID (default %q)", internalapp.DefaultStackID))
	cmd.Flags().StringVar(&flags.projectType, "project-type", "", fmt.Sprintf("project type for server-side preset (default %q)", internalapp.DefaultProjectType))
	cmd.Flags().BoolVar(&flags.public, "public", false, "create as a public app")
	cmd.Flags().StringVar(&flags.bitriseYML, "bitrise-yml", "", "path to bitrise.yml to upload (default: ./bitrise.yml if present, else skip)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("provider", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "github", "gitlab", "bitbucket", "custom"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("project-type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"ios", "android", "flutter", "react-native", "xamarin", "cordova", "ionic", "other"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runCreate(cmd *cobra.Command, detector internalapp.GitDetector, flags createFlags) error {
	cmdutil.LogCommandParameters(cmd)

	format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
	if err := output.ConfigureOutputFormat(format); err != nil {
		return fmt.Errorf("failed to configure output format: %w", err)
	}

	ctx := cmd.Context()
	resolvedRepoURL, gitURLDetected, err := resolveRepoURL(ctx, flags.repoURL, detector)
	if err != nil {
		return err
	}
	resolvedBranch, gitBranchDetected, err := resolveBranch(ctx, flags.branch, detector)
	if err != nil {
		return err
	}
	yamlContent, yamlSource, err := resolveBitriseYML(flags.bitriseYML)
	if err != nil {
		return err
	}

	// --workspace wins; otherwise fall back to BITRISE_WORKSPACE_ID /
	// default_workspace_id. Still empty means Service.Create auto-detects,
	// which only succeeds when the account has exactly one workspace.
	resolvedWorkspace := flags.workspace
	workspaceFromDefault := false
	if resolvedWorkspace == "" {
		resolvedWorkspace = cmdutil.DefaultWorkspaceSlug(cmd)
		workspaceFromDefault = resolvedWorkspace != ""
	}

	stderr := cmd.ErrOrStderr()
	// These breadcrumbs explain what the command inferred. They're suppressed
	// for machine-readable output so a script's stderr stays clean; the two
	// notices in persistAppSlug are not, since they report a file this command
	// changed and a setting that will override it.
	if output.Format == output.FormatRaw {
		if gitURLDetected {
			fmt.Fprintf(stderr, "Detected repo URL from git: %s\n", resolvedRepoURL)
		}
		if gitBranchDetected {
			fmt.Fprintf(stderr, "Detected branch from git: %s\n", resolvedBranch)
		}
		if workspaceFromDefault {
			fmt.Fprintf(stderr, "Using default workspace: %s\n", resolvedWorkspace)
		}
		switch yamlSource {
		case yamlSourceFlag:
			fmt.Fprintf(stderr, "Uploading bitrise.yml from %s\n", flags.bitriseYML)
		case yamlSourceCwd:
			fmt.Fprintln(stderr, "Uploading bitrise.yml from ./bitrise.yml")
		case yamlSourceNone:
			presetProjectType := flags.projectType
			if presetProjectType == "" {
				presetProjectType = internalapp.DefaultProjectType
			}
			fmt.Fprintf(stderr, "No bitrise.yml provided — using server preset for project-type=%s\n", presetProjectType)
		}
	}

	client, err := cmdutil.NewAPIClient(cmd)
	if err != nil {
		return err
	}

	res, err := internalapp.NewService(client).Create(ctx, internalapp.CreateOptions{
		RepoURL:     resolvedRepoURL,
		Branch:      resolvedBranch,
		Title:       flags.title,
		Provider:    flags.provider,
		OrgSlug:     resolvedWorkspace,
		StackID:     flags.stackID,
		ProjectType: flags.projectType,
		Public:      flags.public,
		BitriseYML:  yamlContent,
	})
	if err != nil {
		return err
	}

	// Print before persisting: the app is already created server-side at this
	// point, so a failure to save app_id locally (read-only config dir, lock
	// timeout) must not swallow the slug and build trigger token the user
	// needs to recover manually.
	if err := output.Render(cmd.OutOrStdout(), output.Format, res, printCreateText); err != nil {
		return err
	}

	if err := persistAppSlug(stderr, res.Slug); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to save app_id: %v\n", err)
	}
	return nil
}

// resolveRepoURL returns the explicit flag value or the cwd's git origin.
// The second return is true when the value came from git detection.
func resolveRepoURL(ctx context.Context, flagValue string, d internalapp.GitDetector) (string, bool, error) {
	if flagValue != "" {
		return flagValue, false, nil
	}
	v, err := d.RemoteURL(ctx)
	if err != nil {
		return "", false, fmt.Errorf("detect git remote: %w", err)
	}
	if v == "" {
		return "", false, errors.New("--repo-url is required (no git origin detected in this directory)")
	}
	return v, true, nil
}

// resolveBranch returns the explicit flag value or the cwd's current branch.
// An empty detection result is not an error — Service.Create falls back to
// DefaultBranchFallback.
func resolveBranch(ctx context.Context, flagValue string, d internalapp.GitDetector) (string, bool, error) {
	if flagValue != "" {
		return flagValue, false, nil
	}
	v, err := d.CurrentBranch(ctx)
	if err != nil {
		return "", false, fmt.Errorf("detect git branch: %w", err)
	}
	return v, v != "", nil
}

type yamlSource int

const (
	yamlSourceNone yamlSource = iota
	yamlSourceFlag
	yamlSourceCwd
)

// resolveBitriseYML reads the bitrise.yml to upload: an explicit --bitrise-yml
// path, else ./bitrise.yml if present, else none.
func resolveBitriseYML(flagPath string) (string, yamlSource, error) {
	if flagPath != "" {
		data, err := os.ReadFile(flagPath)
		if err != nil {
			return "", yamlSourceNone, fmt.Errorf("read %s: %w", flagPath, err)
		}
		return string(data), yamlSourceFlag, nil
	}
	const cwdPath = "bitrise.yml"
	data, err := os.ReadFile(cwdPath)
	if errors.Is(err, fs.ErrNotExist) {
		return "", yamlSourceNone, nil
	}
	if err != nil {
		return "", yamlSourceNone, fmt.Errorf("read %s: %w", cwdPath, err)
	}
	return string(data), yamlSourceCwd, nil
}

// persistAppSlug saves slug as the global default app_id and warns on stderr
// if a per-directory .bitrise-cli.yml will still override it at runtime.
func persistAppSlug(stderr io.Writer, slug string) error {
	if err := configs.SetAppID(slug); err != nil {
		return fmt.Errorf("save app_id: %w", err)
	}
	cfgPath, _ := internalconfig.Path()
	fmt.Fprintf(stderr, "Set app_id=%s in %s\n", slug, cfgPath)

	dirCfg, dirPath, err := internalconfig.LoadDir()
	if err == nil && dirPath != "" && dirCfg.AppID != "" && dirCfg.AppID != slug {
		fmt.Fprintf(stderr, "Note: %s pins app_id=%s, which still wins at runtime\n", dirPath, dirCfg.AppID)
	}
	return nil
}

func printCreateText(w io.Writer, r internalapp.CreateResult) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Created app %s\n", r.Slug)
	fmt.Fprintf(&b, "Title:               %s\n", r.Title)
	fmt.Fprintf(&b, "Repo URL:            %s\n", r.RepoURL)
	fmt.Fprintf(&b, "Default branch:      %s\n", r.DefaultBranch)
	fmt.Fprintf(&b, "Workspace:           %s\n", r.OrgSlug)
	fmt.Fprintf(&b, "Stack:               %s\n", r.StackID)
	fmt.Fprintf(&b, "Project type:        %s\n", r.ProjectType)
	if !r.BitriseYMLUploaded {
		fmt.Fprintln(&b, "Config:              server preset")
	}
	// Masked like 'auth status' does, so the token doesn't land in a shell
	// scrollback or CI log by default; --format json still carries the value.
	if r.BuildTriggerToken != "" {
		fmt.Fprintln(&b, "Build trigger token: ******** (set, use --format json or yml to read it)")
	}
	_, err := io.WriteString(w, b.String())
	return err
}
