## bitrise api

Make an authenticated request to the Bitrise API

### Synopsis

Make an authenticated HTTP request to the Bitrise API and print the response.

PATH is resolved relative to the configured API base URL, or used verbatim
if it's an absolute http(s):// URL. The response body goes to stdout, so it
composes with tools like jq.

--field pairs are sent as strings; use --input for bodies that need nesting
or non-string values.

Unlike every other cloud command, this one has no --format flag, so -f is
--field here instead of --format.

```
bitrise api PATH [flags]
```

### Examples

```
  bitrise api /me
  bitrise api /apps -f sort_by=last_build_at --all | jq '.data[].title'
  bitrise api "/apps/APP_ID/builds?limit=10"
  bitrise api /apps/APP_ID/builds -X POST --input body.json
  bitrise api -X DELETE /apps/APP_ID/builds/BUILD_ID -i
```

### Options

```
      --all                  follow cursor pagination and merge every page's data array (GET only)
  -f, --field stringArray    key=value; query parameter for GET, JSON body field otherwise (may be repeated)
  -H, --header stringArray   "Name: value"; adds or overrides a request header (may be repeated)
  -h, --help                 help for api
  -i, --include              print the response status and headers before the body
      --input string         path to a file for the request body, or - for stdin
  -X, --method string        HTTP method (default GET, or POST when a body is present)
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

* [bitrise](bitrise.md)	 - Bitrise Automations Workflow Runner

