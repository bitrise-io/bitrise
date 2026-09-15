## bitrise rde template update

Update an existing RDE template from a JSON spec file

### Synopsis

Update an existing RDE template from a JSON spec file.

Only fields present in the file are sent. Array fields (template_variables,
session_inputs, feature_flags, workspace_links) replace the server's
existing list wholesale when present — to clear one, include it as [].

Note: 'template view' output doesn't include a workspace link's
feature_flag_name (the API doesn't return it), so the round-trip workflow
below silently drops it from any existing links — reapply it by hand.

The template's declared device (the one sessions created from it boot
unless overridden) is replaced when device_spec is present in the file or
when --device-platform is given (with optional --device-model,
--device-os-version, --device-system-image — the same flags 'rde session
create' takes; they take precedence over the file). Pass --clear-device to
remove it. Either may be used without --file. Read 'bitrise rde
device-guide' before declaring one.

Round-trip workflow:

  bitrise rde template view TEMPLATE_ID --format json > template.json
  # edit template.json
  bitrise rde template update TEMPLATE_ID --file template.json

Pass --file - to read the JSON from stdin.

```
bitrise rde template update TEMPLATE_ID [flags]
```

### Examples

```
  bitrise rde template update TEMPLATE_ID --file template.json
  # Declare (or replace) the device sessions from this template boot by default.
  bitrise rde template update TEMPLATE_ID --device-platform android --device-model pixel_7
  # Stop declaring a device.
  bitrise rde template update TEMPLATE_ID --clear-device
```

### Options

```
      --clear-device                 remove the template's declared device so sessions created from it boot none
      --device-model string          device to boot: simctl device type ("iPhone 16") or emulator device profile ("pixel_7"); default: platform default; requires --device-platform
      --device-os-version string     iOS only: an iOS version ("18.2") or simctl runtime id — anything else is rejected; default: newest installed; requires --device-platform
      --device-platform string       declare a virtual device sessions created from the template boot unless overridden: ios (simulator, macOS stack) or android (emulator, Linux stack); read 'rde device-guide' first
      --device-system-image string   Android only: sdkmanager system image package ("system-images;android-34;google_apis;x86_64"); default: platform default; requires --device-platform
  -f, --file string                  path to a JSON spec file (use '-' for stdin); optional when only changing the device
      --format string                Output format. Accepted: raw (default), json, yml
  -h, --help                         help for update
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

* [bitrise rde template](bitrise_rde_template.md)	 - List and inspect RDE templates

