## bitrise local tools setup

Install tools from version files or bitrise.yml

### Synopsis

Install tools from version files (e.g. .tool-versions, .node-version, .nvmrc, .fvmrc, .fvm/fvm_config.json, package.json) or bitrise.yml.

EXAMPLES:
   bitrise tools setup --config .tool-versions
   bitrise tools setup --config .nvmrc
   bitrise tools setup --config .fvmrc
   bitrise tools setup --config .fvm/fvm_config.json
   bitrise tools setup --config package.json
   bitrise tools setup --config bitrise.yml
   bitrise tools setup --provider mise --fast-install true

   Setup and activate in current shell session:
   eval "$(bitrise tools setup --config .tool-versions --format bash)"

```
bitrise local tools setup [--provider PROVIDER] [--fast-install true|false] [--config FILE] [--format FORMAT] [--workflow WORKFLOW] [flags]
```

### Options

```
  -c, --config stringArray    Config or version file paths to install tools from. Can be specified multiple times. If not provided, detects files in the working directory. Supported file names and formats:
                              	- .tool-versions (asdf/mise style): multiple tools, one "<tool> <version>" per line
                              	- .<tool>-version (e.g. .node-version, .ruby-version): single tool, version string only
                              	- .nvmrc (NVM): Node.js version
                              	- .fvmrc (FVM 3.x): Flutter version from JSON {"flutter": "<version>"}
                              	- .fvm/fvm_config.json (legacy FVM): Flutter version from {"flutterSdkVersion": "<version>"}
                              	- package.json: Node.js version from the "engines.node" field
                              	- bitrise.yml: tools defined in the "tools" section
      --fast-install string   Override fast install setting (true/false). Fast install uses Lix (Nix) for faster installation. If not specified, uses the default for the used stack.
  -f, --format string         Output format of the env vars that activate installed tools. Options: plaintext, json, bash (default "plaintext")
  -h, --help                  help for setup
  -p, --provider string       Tool provider to use (asdf/mise). If not specified, uses the default.
  -w, --workflow string       Workflow ID to use when installing from bitrise.yml (optional, uses global tools if not specified)
```

### Options inherited from parent commands

```
      --ci              If true it indicates that we're used by another tool so don't require any user input!
      --debug           If true it enables DEBUG mode.
      --no-color        Disable ANSI colors (the NO_COLOR env var is also honored).
  -o, --output string   Output format for commands that support it. Accepted: raw (default), json, yml (alias "human").
      --pr              If true bitrise runs in pull request mode.
  -q, --quiet           Suppress non-error diagnostic messages.
      --theme string    Color theme. Accepted: auto, dark, light, none.
```

### SEE ALSO

* [bitrise local tools](bitrise_local_tools.md)	 - Manage available tools from inside the workflow.

