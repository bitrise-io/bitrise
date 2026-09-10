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
```

### Options

```
      --ai-prompt string             initial AI prompt passed to Claude Code on session start
      --auto-terminate-minutes int   minutes until auto-termination; 0 disables; omitted uses backend default (~5 days)
      --cluster string               target cluster name (use 'rde machine-type list --stack STACK_ID' to find candidates when the stack + machine type combo is ambiguous)
      --description string           session description
      --feature-flag stringArray     name of a feature flag to enable on the session (repeatable)
  -f, --format string                Output format. Accepted: raw (default), json, yml
  -h, --help                         help for create
      --input stringArray            session input as key=value (repeatable)
  -l, --label stringArray            label to attach to the session as key=value (repeatable; at most 32; keys use letters, digits, and . _ / -, values additionally : and +; the bitrise.io/ key prefix is reserved)
      --machine-type string          machine type name for a template-less session, or to override the template's machine type (see 'rde machine-type list --stack STACK_ID')
      --map-saved-inputs             auto-fill template session inputs from the user's saved inputs (matched by key)
      --saved-input stringArray      session input as key=savedInputID — uses a stored saved-input value (repeatable)
      --secret-input stringArray     session input as key=value, stored as a secret at rest (repeatable; the value is visible in shell history and process args — prefer --saved-input)
      --stack string                 stack ID for a template-less session, or to override the template's stack (see 'rde stack list')
      --template string              template ID or name to create the session from (omit to create a template-less session with --stack and --machine-type)
      --wait                         wait until the session leaves provisioning (running, failed, …) before returning; exits 1 if the final status isn't running
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

