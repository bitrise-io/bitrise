## bitrise user create

Create a new Bitrise account

### Synopsis

Create a new Bitrise account by email and password.

Password input:
  By default the command prompts for the password (input is masked when stdin
  is a terminal). Use --password-stdin to read it from stdin without a prompt
  — the right choice for piping or scripts:

      printf '%s' "$NEW_PASSWORD" | bitrise user create \
          --email a@b.io --username alice --first-name A --last-name B --password-stdin

Email verification:
  After signup the server emails a verification link. Click it before running
  'bitrise auth login --email <addr>' — sign-in is blocked on unverified
  accounts.

Target host:
  The signup request goes to web_base_url (default app.bitrise.io),
  overridable via $BITRISE_WEB_BASE_URL or 'bitrise config set web_base_url'
  — never by a per-directory .bitrise-cli.yml, so a repo you merely clone
  can't silently redirect where your password is sent.

```
bitrise user create [flags]
```

### Examples

```
  bitrise user create --email alice@example.com --username alice --first-name Alice --last-name L
  printf '%s' "$NEW_PASSWORD" | bitrise user create \
      --email alice@example.com --username alice --first-name Alice --last-name L --password-stdin --format json
```

### Options

```
      --email string        email address to register (required)
      --first-name string   first name on the account (required)
  -f, --format string       Output format. Accepted: raw (default), json, yml
  -h, --help                help for create
      --last-name string    last name on the account (required)
      --password-stdin      read the password from stdin without prompting
      --username string     desired username (must be unique) (required)
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

* [bitrise user](bitrise_user.md)	 - Create and manage your Bitrise account.

