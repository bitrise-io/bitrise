## bitrise rde session create

Create a new RDE session

### Synopsis

Create a new RDE session, either from a template or from a bare
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
  --secret-input api-key=VALUE               # inline; avoid for real secrets

```
bitrise rde session create NAME [flags]
```

### Examples

```
  bitrise rde session create dev --template TEMPLATE_ID
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
  bitrise rde session create android-check --device-platform android --artifact-url-stdin < artifact-url.txt
```

### Options

```
      --ai-prompt string             initial AI prompt passed to Claude Code on session start
      --artifact-name string         display name of the app installed from --artifact-url / --artifact-url-stdin
      --artifact-url string          app build to install once the device is ready: absolute http(s) URL of a zipped simulator .app (iOS) or an .apk (Android); requires --device-platform or a --template that declares a device (a signed URL is visible in shell history and process args — prefer --artifact-url-stdin)
      --artifact-url-stdin           read the --artifact-url value from stdin instead of the command line; keeps signed URLs out of shell history and process args; requires --device-platform
      --auto-terminate-minutes int   minutes until auto-termination; 0 disables; omitted uses the backend default (~5 days)
      --cluster string               target cluster name (use 'rde machine-type list --stack STACK_ID' to find candidates when the stack + machine type combo is ambiguous)
      --description string           session description
      --device-model string          device to boot: simctl device type ("iPhone 16") or emulator device profile ("pixel_7") — a screen profile, not that phone's firmware; default: the template's device model, else the platform default
      --device-os-version string     iOS only: an iOS version ("18.2") or simctl runtime id — anything else is rejected; default: the template's, else newest installed
      --device-platform string       boot a virtual device with the session: ios (simulator, macOS stack) or android (emulator, Linux stack); --stack/--machine-type may then be omitted; read 'rde device-guide' first
      --device-system-image string   Android only: the API-level knob — sdkmanager system image package ("system-images;android-34;google_apis;x86_64"); default: the template's, else the platform default
      --feature-flag stringArray     name of a feature flag to enable on the session (repeatable)
  -f, --format string                Output format. Accepted: raw (default), json, yml
  -h, --help                         help for create
      --input stringArray            session input as key=value (repeatable)
  -l, --label stringArray            label to attach to the session as key=value (repeatable; at most 32; keys use letters, digits, and . _ / -, values additionally : and +; the bitrise.io/ key prefix is reserved)
      --machine-type string          machine type name for a template-less session, or to override the template's machine type (see 'rde machine-type list --stack STACK_ID')
      --map-saved-inputs             auto-fill template session inputs from the user's saved inputs (matched by key)
      --no-device                    create without the template's device (ignored when the template declares none)
      --saved-input stringArray      session input as key=savedInputID — uses a stored saved-input value (repeatable)
      --secret-input stringArray     session input as key=value, stored as a secret at rest (repeatable; the value is visible in shell history and process args — prefer --saved-input)
      --stack string                 stack ID for a template-less session, or to override the template's stack (see 'rde stack list')
      --template string              template ID or name to create the session from (omit to create a template-less session with --stack and --machine-type)
      --wait                         wait until the session leaves provisioning (running, failed, …) — and, with --device-platform, until the device is ready or failed — before returning; exits 1 if the final status isn't running
      --wait-timeout duration        max time to wait when --wait is set (uses Go duration syntax: 30s, 5m, 1h) (default 10m0s)
```

### Options inherited from parent commands

```
      --ci                 If true it indicates that we're used by another tool so don't require any user input!
      --debug              If true it enables DEBUG mode.
      --no-color           Disable ANSI colors (the NO_COLOR env var is also honored).
  -o, --output string      Output format for commands that support it. Accepted: raw (default), json, yml (alias "human").
      --pr                 If true bitrise runs in pull request mode.
  -q, --quiet              Suppress non-error diagnostic messages.
      --theme string       Color theme. Accepted: auto, dark, light, none.
      --workspace string   workspace ID (or set BITRISE_WORKSPACE_ID or default_workspace_id; auto-detected if you have exactly one workspace)
```

### SEE ALSO

* [bitrise rde session](bitrise_rde_session.md)	 - Create, list, inspect, and manage RDE sessions

