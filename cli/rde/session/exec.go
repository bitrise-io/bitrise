package session

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newExecCmd() *cobra.Command {
	var (
		shellMode bool
		timeout   time.Duration
		envFlags  []string
		noEnvFile bool
		format    string
	)
	c := &cobra.Command{
		Use:   "exec SESSION_ID -- COMMAND [ARGS...]",
		Short: "Run a command on a session over SSH",
		Long: `Run a command on a session over SSH and capture its output.

The command runs in a forced-interactive login bash shell (bash -i -l -c) so
the session's PATH, brew-installed binaries, git-lfs, and language version
managers (nvm, pyenv, rbenv, asdf) are all loaded.

By default the tokens after '--' are treated as a program plus literal
arguments: each is passed through verbatim, so shell metacharacters (;, &&,
|, $(...), redirection) are NOT interpreted — 'exec ID -- echo "a; b"' runs
echo with the single literal argument 'a; b'. This keeps quoted arguments
intact (e.g. -m "a message"). (Quote such arguments so your local shell
hands them over as one token rather than splitting them itself.)

Pass --shell to interpret everything after '--' as a shell command line
instead, so pipes, &&, command substitution and redirection work:

  bitrise rde session exec SESSION_ID --shell -- 'cd repo && xcodebuild | xcpretty'

This is the first-class replacement for hand-wrapping every command in
bash -lc "…". Quote the command (or its metacharacters) so your local shell
passes them through unchanged. --shell must come before '--'; after '--' it
is just another literal token.

If a local SSH agent is available ($SSH_AUTH_SOCK set), it's forwarded into
the session — git-over-SSH inside the session uses the caller's local keys.

Local environment variables can be forwarded to the remote command with
--env (repeatable, before '--'): --env NAME forwards the local value of
NAME (an error if unset), --env NAME=VALUE sets a literal. A repo can pin
a shared list in .bitrise/rde.yml (found in the working directory or any
ancestor):

  exec:
    env:
      - API_BASE_URL
      - NPM_TOKEN=abc123

File entries use the same NAME / NAME=VALUE forms, except a NAME that is
unset locally is skipped with a warning instead of failing, so a shared
file never breaks a teammate. --env overrides a same-named file entry;
--no-env-file skips the file entirely. Commit the file to share the list
with your team, or gitignore it to keep a personal list (including
NAME=VALUE literals) out of the repo. Forwarded variables are announced
by name on stderr (silence with -q) — values are never printed locally,
but they do ride in the remote command line, so they are visible in ps
on the session while the command runs.

In raw mode: stdout streams to this CLI's stdout, stderr to stderr, and
this CLI exits non-zero when the remote command exits non-zero.

In --format json/yml mode: a single exit_code/stdout/stderr object is
emitted to stdout regardless of the command's exit status.

The remote command is capped at 10 minutes by default; raise it with --timeout
(e.g. --timeout 20m for a cold xcodebuild) or pass --timeout 0 to disable the
cap. exec holds the SSH connection open for the whole run, so the command dies
if the connection drops — for fire-and-forget work that must outlive the
connection, nohup it inside the session instead.`,
		Example: `  bitrise rde session exec SESSION_ID -- echo hello
  bitrise rde session exec SESSION_ID -- npm test
  bitrise rde session exec SESSION_ID -- git commit -m "a message"
  bitrise rde session exec SESSION_ID --shell -- 'cd repo && ls | head'
  bitrise rde session exec SESSION_ID --env NPM_TOKEN --env CI=1 -- npm test
  bitrise rde session exec SESSION_ID --timeout 20m -- ./scripts/cold-build.sh
  bitrise rde session exec SESSION_ID --format json -- ls -la /opt`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("session exec requires SESSION_ID followed by a command (e.g. 'rde session exec ID -- echo hi')")
			}
			return nil
		},
		// DisableFlagParsing stays off so our own flags (--shell, --format …)
		// are parsed. Combined with the explicit `--` convention, tokens after
		// `--` are still handed through as positional args, so a user can write
		// `exec ID -- ls -la /opt` without `-la` being read as our flag.
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			sessionID := args[0]
			command := buildExecCommand(args[1:], shellMode)
			if strings.TrimSpace(command) == "" {
				return fmt.Errorf("command is empty")
			}
			// Resolve the forwarded env before anything touches the network,
			// so a typo'd --env or a broken dotfile fails fast.
			var fileEntries []string
			var envFilePath string
			if !noEnvFile {
				repoCfg, path, err := internalrde.LoadRepoConfig()
				if err != nil {
					return err
				}
				fileEntries, envFilePath = repoCfg.Exec.Env, path
			}
			envVars, skippedEnv, err := internalrde.ResolveExecEnv(fileEntries, envFilePath, envFlags, os.LookupEnv)
			if err != nil {
				return err
			}
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)
			sessionID, err = svc.ResolveSessionID(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			// Notices print here — after session resolution, so "Forwarding"
			// only announces an exec that's actually about to run — and
			// before Execute, so they land ahead of any streamed output.
			if err := printExecEnvNotices(cmd, envVars, skippedEnv, envFilePath); err != nil {
				return err
			}
			res, err := svc.Execute(cmd.Context(), workspaceID, sessionID, command, envVars, timeout)
			if err != nil {
				return err
			}
			return renderExecResult(cmd, res)
		},
	}
	c.Flags().BoolVar(&shellMode, "shell", false, "interpret everything after '--' as a shell command line (pipes, &&, $(...), redirection) instead of a program with literal arguments")
	c.Flags().DurationVar(&timeout, "timeout", internalrde.DefaultExecuteTimeout, "max time the remote command may run before it's aborted; 0 disables the cap (Go duration syntax: 30s, 10m, 1h)")
	c.Flags().StringArrayVar(&envFlags, "env", nil, "environment variable for the remote command: NAME forwards the local value (errors if unset), NAME=VALUE sets a literal (repeatable; must come before '--')")
	c.Flags().BoolVar(&noEnvFile, "no-env-file", false, "skip reading forwarded env vars from .bitrise/rde.yml")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

// printExecEnvNotices writes the env-forwarding diagnostics to stderr: a
// warning naming dotfile entries skipped because they're unset locally
// (never suppressed — warnings ignore --quiet), then a names-only forwarding
// notice (suppressed by --quiet). Values are never printed.
func printExecEnvNotices(cmd *cobra.Command, vars []internalrde.EnvVar, skipped []string, filePath string) error {
	ew := cmdutil.NewErrWriter(cmd.ErrOrStderr())
	if len(skipped) > 0 {
		s := style.New(cmd.ErrOrStderr())
		ew.F("%s %s not set locally — skipped (listed in %s)\n", s.Warn.Render("Warning:"), strings.Join(skipped, ", "), filePath)
	}
	if len(vars) > 0 && !cmdutil.IsQuiet(cmd) {
		names := make([]string, len(vars))
		for i, v := range vars {
			names[i] = v.Name
		}
		ew.F("Forwarding env: %s\n", strings.Join(names, ", "))
	}
	return ew.Err
}

// buildExecCommand turns the post-`--` tokens into the command string sent to
// the session. The remote side always wraps it in `bash -i -l -c '<cmd>'`
// (see internal/rde), so the only question here is how the tokens are joined:
//
//   - shell mode: joined verbatim with spaces, so the wrapped bash interprets
//     metacharacters (matches the backend's `bash_c_command`). The caller is
//     responsible for quoting so their local shell passes the metacharacters
//     through.
//   - default: each token is POSIX-quoted, so it reaches bash as a literal
//     argv element and shell metacharacters are inert.
func buildExecCommand(cmdArgs []string, shell bool) string {
	if shell {
		return strings.Join(cmdArgs, " ")
	}
	return cmdutil.JoinShellArgs(cmdArgs)
}

func renderExecResult(cmd *cobra.Command, res internalrde.ExecResult) error {
	if output.Format != output.FormatRaw {
		return output.Print(cmd.OutOrStdout(), res, output.Format)
	}
	// Raw mode: stream stdout and stderr to their natural sinks. We don't
	// echo a JSON-shaped envelope — users piping `| jq` should be using
	// --format json.
	if res.Stdout != "" {
		if _, err := io.WriteString(cmd.OutOrStdout(), res.Stdout); err != nil {
			return err
		}
	}
	if res.Stderr != "" {
		if _, err := io.WriteString(cmd.ErrOrStderr(), res.Stderr); err != nil {
			return err
		}
	}
	if res.ExitCode != 0 {
		// Suppress cobra's "Error: ..." line — the remote command's
		// stderr already explains the failure.
		cmdutil.SilenceRootErrors(cmd)
		return fmt.Errorf("remote command exited with status %d", res.ExitCode)
	}
	return nil
}
