package session

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newCreateCmd() *cobra.Command {
	var (
		description          string
		templateID           string
		stack                string
		machineType          string
		inputs               []string
		secretInputs         []string
		savedInputs          []string
		labels               []string
		featureFlags         []string
		cluster              string
		aiPrompt             string
		autoTerminateMinutes int
		setAutoTerminate     bool
		mapSavedInputs       bool
		wait                 bool
		waitTimeout          time.Duration
		format               string
		devicePlatform       string
		deviceModel          string
		deviceOSVersion      string
		deviceSystemImage    string
		noDevice             bool
		artifactURL          string
		artifactURLStdin     bool
		artifactName         string
	)

	c := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new RDE session",
		Long: `Create a new RDE session, either from a template or from a bare
stack + machine type (a template-less session, with no warmup/startup scripts
or other template configuration).

NAME is a human-readable label for the session; you can use it in place of the
session ID in later commands (view, terminate, …) as long as it stays unique.

Pass --template to create the session from a template (by ID or name). To
create a session without a template, omit --template and pass both --stack and
--machine-type instead. --stack / --machine-type may also be given alongside
--template to override the template's defaults for this session.

Provide session input values via --input (one --input per key), --secret-input
(value stored as secret-at-rest), or --saved-input (reference an existing saved
input by ID). Use --map-saved-inputs to auto-fill any session input key that
matches a saved input the user already has.

For secret values, prefer storing them once with 'rde saved-input create
--value-stdin --secret' and referencing them by ID via --saved-input. A value
passed inline with --secret-input ends up in your shell history and in the
process arguments (readable by other users via 'ps'); marking it secret only
governs how the backend stores the value, not how it reaches the CLI.

Attach arbitrary key=value metadata with --label (repeatable); labels come
back on 'session view' and in 'session list --format json', and sessions
can be filtered by them with 'rde session list --label-selector key=value'.

Want a device on the session? READ THE GUIDE FIRST: 'bitrise rde
device-guide' (then 'rde device-guide ios' or 'android' for the platform you
boot). It covers the readiness contract, connecting, driving the device
efficiently, letting a human watch, recovery, and what never to do.

Boot a virtual device with the session by passing --device-platform ios (an
iOS simulator on a macOS stack) or android (an Android emulator on a dockerless
Android Linux stack). Prefer omitting --stack/--machine-type: the deployment's
known-good pair for the platform applies (the guide says which stacks fit when
you must name one); with --template the template's stack and machine type are
used and must fit the platform. --cluster is never needed with a device.
--device-model is a screen profile (size, density), not that phone's
firmware; --device-system-image picks the Android API level.

A template may declare a device of its own ('rde template view' shows it as
"Device:"). Sessions created from such a template boot that device as
declared — no device flags needed. The template's device is the base: with
--template, --device-model, --device-os-version and --device-system-image may
be given without --device-platform and tweak the template's device per field
(unset fields inherit the template's). Passing --device-platform makes the
flags the complete device to boot: the template's is ignored and unset fields
are the platform defaults. Pass --no-device to create the session without the
template's device. Optionally pre-install an app with --artifact-url, or
--artifact-url-stdin to read the URL from stdin: a signed (pre-authenticated)
download URL is a bearer credential, and a value passed inline ends up in
your shell history and in the process arguments (readable by other users via
'ps'). "running" does not mean the device is usable — 'session view' shows
the device state; wait for "ready" (--wait does so for you when a device was
requested) and touch nothing on the VM while it is "booting". A "failed"
device is not always unusable: 'session view' prints the device notes, and
the guide says which failures leave the device drivable.

Example values:
  --input key=value
  --saved-input session-key=SAVED_INPUT_ID   # secret stored ahead of time
  --secret-input api-key=VALUE               # inline; avoid for real secrets`,
		Example: `  bitrise rde session create dev --template TEMPLATE_ID
  bitrise rde session create dev --template TEMPLATE_ID --input repo=my-app
  # Template-less: pick a stack and machine type directly.
  bitrise rde session create dev --stack osx-xcode-16.0.x-edge --machine-type g2.mac.m2pro.6c-14g
  # Keep secrets off the command line: store once, then reference by ID.
  echo -n "ghp_xxx" | bitrise rde saved-input create --key gh-token --value-stdin --secret
  bitrise rde session create dev --template TEMPLATE_ID --saved-input gh-token=SAVED_INPUT_ID
  bitrise rde session create dev --template TEMPLATE_ID --map-saved-inputs
  # Boot an iOS simulator with the session (stack/machine type default to the platform's).
  bitrise rde device-guide ios     # read first: readiness, connecting, driving, do-nots
  bitrise rde session create ios-check --device-platform ios --device-model "iPhone 16" --device-os-version 18.2
  bitrise rde session create android-check --device-platform android --artifact-url https://…/app.apk
  # From a template that declares a device: boot it as declared, override one field, or skip it.
  bitrise rde session create ios-check --template TEMPLATE_ID
  bitrise rde session create ios-check --template TEMPLATE_ID --device-model "iPhone 15"
  bitrise rde session create no-sim --template TEMPLATE_ID --no-device
  # Keep a signed artifact URL out of shell history and process args: read it from a file.
  bitrise rde session create android-check --device-platform android --artifact-url-stdin < artifact-url.txt`,
		Args: cmdutil.RequireArgs("NAME"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			name := args[0]
			if name == "" {
				return fmt.Errorf("NAME must not be empty")
			}
			// A session needs either a template or, for a template-less
			// session, an explicit stack + machine type — unless a device is
			// requested, in which case the backend fills whatever is missing
			// from the platform's defaults. (stack/machine type may also
			// accompany a template to override its defaults.)
			if templateID == "" && devicePlatform == "" && (stack == "" || machineType == "") {
				return fmt.Errorf("provide --template, --device-platform, or both --stack and --machine-type to create a session without a template")
			}
			switch devicePlatform {
			case "", "ios", "android":
			default:
				return fmt.Errorf("--device-platform must be ios or android")
			}
			// The per-device fields need a platform to attach to — unless a
			// template supplies it: then they tweak the template's declared
			// device per field (the platform, and unset fields, inherit).
			// With --device-platform the flags are the complete device.
			deviceFieldsSet := deviceModel != "" || deviceOSVersion != "" || deviceSystemImage != ""
			if devicePlatform == "" && templateID == "" && deviceFieldsSet {
				return fmt.Errorf("--device-model, --device-os-version and --device-system-image require --device-platform (or --template with a declared device)")
			}
			// An artifact needs a device to land on: one requested here, or the
			// one a template declares (the backend rejects a template without one).
			if devicePlatform == "" && templateID == "" && (artifactURL != "" || artifactURLStdin || artifactName != "") {
				return fmt.Errorf("--artifact-url, --artifact-url-stdin and --artifact-name require --device-platform (or --template with a template that declares a device)")
			}
			if artifactURLStdin {
				// Signed download URLs are bearer credentials; reading them
				// from stdin keeps them out of shell history and `ps` (same
				// rationale as `saved-input create --value-stdin`).
				v, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "", true)
				if err != nil {
					return fmt.Errorf("reading --artifact-url-stdin: %w", err)
				}
				if v == "" {
					return fmt.Errorf("--artifact-url-stdin: no URL read from stdin")
				}
				artifactURL = v
			}
			if devicePlatform == "ios" && deviceSystemImage != "" {
				return fmt.Errorf("--device-system-image applies to Android only")
			}
			if artifactName != "" && artifactURL == "" {
				return fmt.Errorf("--artifact-name requires --artifact-url or --artifact-url-stdin")
			}
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			sessionInputs, err := parseSessionInputs(inputs, secretInputs, savedInputs)
			if err != nil {
				return err
			}
			labelMap, err := parseLabelFlags("--label", labels)
			if err != nil {
				return err
			}
			req := internalrde.CreateSessionRequest{
				Name:                    name,
				Description:             description,
				TemplateID:              templateID,
				StackID:                 stack,
				MachineType:             machineType,
				SessionInputs:           sessionInputs,
				EnabledFeatureFlagNames: featureFlags,
				Cluster:                 cluster,
				AIPrompt:                aiPrompt,
				MapSavedToSessionInputs: mapSavedInputs,
				Labels:                  labelMap,
			}
			if setAutoTerminate {
				m := autoTerminateMinutes
				req.AutoTerminateMinutes = &m
			}
			if devicePlatform != "" || deviceFieldsSet {
				// An empty Platform is deliberate: with --template the
				// backend merges this spec over the template's device and
				// fills the platform (and any other unset field) from it.
				req.DeviceSpec = &internalrde.DeviceSpec{
					Platform:    devicePlatform,
					DeviceModel: deviceModel,
					OSVersion:   deviceOSVersion,
					SystemImage: deviceSystemImage,
				}
				if artifactURL != "" {
					req.Artifact = &internalrde.DeviceArtifact{URL: artifactURL, AppName: artifactName}
				}
			}
			req.NoDevice = noDevice
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)

			// --template accepts either a UUID or a template name; resolve
			// names to IDs before issuing CreateSession so the user gets
			// a clean error if the name is wrong or ambiguous. Skipped for
			// template-less sessions, where no template is involved.
			if req.TemplateID != "" {
				resolvedID, err := svc.ResolveTemplateID(cmd.Context(), workspaceID, req.TemplateID)
				if err != nil {
					return err
				}
				req.TemplateID = resolvedID
			}

			res, err := svc.CreateSession(cmd.Context(), workspaceID, req)
			if err != nil {
				return err
			}

			if wait {
				waitCtx, cancel := context.WithTimeout(cmd.Context(), waitTimeout)
				defer cancel()
				if !cmdutil.IsQuiet(cmd) && output.Format == output.FormatRaw {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Waiting for session %s to become ready (timeout %s)…\n", res.Session.ID, waitTimeout)
				}
				ready, waitErr := svc.WaitForReady(waitCtx, workspaceID, res.Session.ID, 0, nil)
				if waitErr != nil {
					// The session exists and is billing even though the wait
					// failed; render it so its ID isn't lost — the only other
					// place it appears is the "Waiting for session …"
					// breadcrumb, suppressed under --quiet and any non-raw
					// format.
					if renderErr := output.Render(cmd.OutOrStdout(), output.Format, res, renderCreateResult); renderErr != nil {
						return renderErr
					}
					return fmt.Errorf("waiting for session: %w", waitErr)
				}
				// A running VM is not a usable device: when one was requested,
				// keep polling (same timeout budget) until it reports ready or failed.
				for ready.Status == "running" && ready.Device != nil && (ready.Device.State == "" || ready.Device.State == "booting") {
					select {
					case <-waitCtx.Done():
						return fmt.Errorf("waiting for device: %w", waitCtx.Err())
					case <-time.After(deviceWaitPollInterval):
					}
					if ready, waitErr = svc.GetSession(waitCtx, workspaceID, res.Session.ID); waitErr != nil {
						return fmt.Errorf("waiting for device: %w", waitErr)
					}
				}
				res.Session = ready
				// A READY device with notes is a substituted device (an
				// uninstalled system image or iOS runtime, clamped sizing):
				// the device spec keeps echoing the request, so say it loudly
				// where the caller is looking, in both output modes.
				if ready.Device != nil && ready.Device.State == "ready" && ready.Device.DeviceNotes != "" && !cmdutil.IsQuiet(cmd) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: the device booted with notes — what you got differs from what you asked for: %s\n", ready.Device.DeviceNotes)
				}
				deviceFailed := ready.Device != nil && ready.Device.State == "failed"
				if ready.Status != "running" || deviceFailed {
					if renderErr := output.Render(cmd.OutOrStdout(), output.Format, res, renderCreateResult); renderErr != nil {
						return renderErr
					}
					cmdutil.SilenceRootErrors(cmd)
					if deviceFailed {
						return fmt.Errorf("session is running but its device failed to boot: %s", ready.Device.DeviceNotes)
					}
					return fmt.Errorf("session ended provisioning with status %q (expected running)", ready.Status)
				}
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, renderCreateResult)
		},
	}

	c.Flags().StringVar(&description, "description", "", "session description")
	c.Flags().StringVar(&templateID, "template", "", "template ID or name to create the session from (omit to create a template-less session with --stack and --machine-type)")
	c.Flags().StringVar(&stack, "stack", "", "stack ID for a template-less session, or to override the template's stack (see 'rde stack list')")
	c.Flags().StringVar(&machineType, "machine-type", "", "machine type name for a template-less session, or to override the template's machine type (see 'rde machine-type list --stack STACK_ID')")
	c.Flags().StringArrayVar(&inputs, "input", nil, "session input as key=value (repeatable)")
	c.Flags().StringArrayVar(&secretInputs, "secret-input", nil, "session input as key=value, stored as a secret at rest (repeatable; the value is visible in shell history and process args — prefer --saved-input)")
	c.Flags().StringArrayVar(&savedInputs, "saved-input", nil, "session input as key=savedInputID — uses a stored saved-input value (repeatable)")
	c.Flags().StringArrayVarP(&labels, "label", "l", nil, "label to attach to the session as key=value (repeatable; at most 32; keys use letters, digits, and . _ / -, values additionally : and +; the bitrise.io/ key prefix is reserved)")
	c.Flags().StringArrayVar(&featureFlags, "feature-flag", nil, "name of a feature flag to enable on the session (repeatable)")
	c.Flags().StringVar(&cluster, "cluster", "", "target cluster name (use 'rde machine-type list --stack STACK_ID' to find candidates when the stack + machine type combo is ambiguous)")
	c.Flags().StringVar(&aiPrompt, "ai-prompt", "", "initial AI prompt passed to Claude Code on session start")
	c.Flags().IntVar(&autoTerminateMinutes, "auto-terminate-minutes", 0, "minutes until auto-termination; 0 disables; omitted uses the backend default (~5 days)")
	c.Flags().BoolVar(&mapSavedInputs, "map-saved-inputs", false, "auto-fill template session inputs from the user's saved inputs (matched by key)")
	c.Flags().StringVar(&devicePlatform, "device-platform", "", "boot a virtual device with the session: ios (simulator, macOS stack) or android (emulator, Linux stack); --stack/--machine-type may then be omitted; read 'rde device-guide' first")
	c.Flags().StringVar(&deviceModel, "device-model", "", "device to boot: simctl device type (\"iPhone 16\") or emulator device profile (\"pixel_7\") — a screen profile, not that phone's firmware; default: the template's device model, else the platform default")
	c.Flags().StringVar(&deviceOSVersion, "device-os-version", "", "iOS only: an iOS version (\"18.2\") or simctl runtime id — anything else is rejected; default: the template's, else newest installed")
	c.Flags().StringVar(&deviceSystemImage, "device-system-image", "", "Android only: the API-level knob — sdkmanager system image package (\"system-images;android-34;google_apis;x86_64\"); default: the template's, else the platform default")
	c.Flags().BoolVar(&noDevice, "no-device", false, "create without the template's device (ignored when the template declares none)")
	c.Flags().StringVar(&artifactURL, "artifact-url", "", "app build to install once the device is ready: absolute http(s) URL of a zipped simulator .app (iOS) or an .apk (Android); requires --device-platform or a --template that declares a device (a signed URL is visible in shell history and process args — prefer --artifact-url-stdin)")
	c.Flags().BoolVar(&artifactURLStdin, "artifact-url-stdin", false, "read the --artifact-url value from stdin instead of the command line; keeps signed URLs out of shell history and process args; requires --device-platform")
	c.Flags().StringVar(&artifactName, "artifact-name", "", "display name of the app installed from --artifact-url / --artifact-url-stdin")
	c.Flags().BoolVar(&wait, "wait", false, "wait until the session leaves provisioning (running, failed, …) — and, with --device-platform, until the device is ready or failed — before returning; exits 1 if the final status isn't running")
	c.Flags().DurationVar(&waitTimeout, "wait-timeout", 10*time.Minute, "max time to wait when --wait is set (uses Go duration syntax: 30s, 5m, 1h)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	c.MarkFlagsMutuallyExclusive("artifact-url", "artifact-url-stdin")
	for _, f := range []string{"device-platform", "device-model", "device-os-version", "device-system-image", "artifact-url", "artifact-url-stdin", "artifact-name"} {
		c.MarkFlagsMutuallyExclusive("no-device", f)
	}

	c.PreRun = func(cmd *cobra.Command, _ []string) {
		// Track whether --auto-terminate-minutes was explicitly set so we
		// can distinguish "not provided" from "set to 0".
		setAutoTerminate = cmd.Flags().Changed("auto-terminate-minutes")
	}
	return c
}

// deviceWaitPollInterval is how often --wait re-reads a device session while
// its device is still booting (a variable so tests can shorten it).
var deviceWaitPollInterval = 3 * time.Second

// parseSessionInputs converts the user-friendly --input/--secret-input/--saved-input
// flags into SessionInputValue entries. Returns an error on the first malformed
// entry; later iterations don't run.
func parseSessionInputs(plain, secret, saved []string) ([]internalrde.SessionInputValue, error) {
	out := make([]internalrde.SessionInputValue, 0, len(plain)+len(secret)+len(saved))
	for _, kv := range plain {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("--input %q: expected key=value", kv)
		}
		out = append(out, internalrde.SessionInputValue{Key: k, Value: v})
	}
	for _, kv := range secret {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("--secret-input %q: expected key=value", kv)
		}
		out = append(out, internalrde.SessionInputValue{Key: k, Value: v, IsSecret: true})
	}
	for _, kv := range saved {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || k == "" || v == "" {
			return nil, fmt.Errorf("--saved-input %q: expected key=savedInputID", kv)
		}
		out = append(out, internalrde.SessionInputValue{Key: k, SavedInputID: v})
	}
	return out, nil
}

func renderCreateResult(w io.Writer, res internalrde.CreateSessionResult) error {
	s := style.New(w)
	ew := cmdutil.NewErrWriter(w)
	ew.F("%s %s\n", s.BuildStatus("success").Render("✓"), "Session created")
	if err := renderSessionDetail(w, res.Session); err != nil {
		return err
	}
	if d := res.Session.Device; d != nil {
		guide := "bitrise rde device-guide"
		if d.Spec != nil && (d.Spec.Platform == "ios" || d.Spec.Platform == "android") {
			guide += " " + d.Spec.Platform
		}
		ew.Ln()
		ew.Ln(s.Dim.Render(fmt.Sprintf("Next: wait until 'rde session view %s' shows the device ready, then drive it per '%s'.", res.Session.ID, guide)))
	}
	if len(res.AutoMappedInputs) > 0 {
		ew.Ln()
		ew.Ln(s.Dim.Render("Auto-mapped session inputs from saved inputs:"))
		for _, m := range res.AutoMappedInputs {
			ew.F("  %s → %s\n", m.SessionInputKey, s.Slug.Render(m.SavedInputID))
		}
	}
	return ew.Err
}
