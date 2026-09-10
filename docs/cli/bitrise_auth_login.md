## bitrise auth login

Save a Bitrise access token

### Synopsis

Save a Bitrise access token for future commands to use.

By default, in an interactive terminal, this opens your browser to sign in to
Bitrise (OAuth) and stores a managed, auto-refreshing token. The modes:

  Browser sign-in (default in an interactive terminal; explicit with --oauth).
     Opens your browser to sign in, exchanges the result for a Personal Access
     Token, and refreshes it automatically so you rarely sign in again:

         bitrise auth login
         bitrise auth login --oauth

     This needs the browser on the same machine as the CLI (the sign-in is
     handed back over a loopback address). On a remote/headless host over SSH
     it can't complete — pipe a token instead (see below).

  Token (--with-token, or any non-interactive stdin).
     Reads a Personal Access Token from stdin, or prompts for one (masked, not
     echoed) when stdin is an interactive terminal. This mode is also used
     automatically when stdin is not a terminal, so CI and pipes keep working
     without a flag:

         echo "$BITRISE_PAT" | bitrise auth login
         echo "$BITRISE_PAT" | bitrise auth login --with-token

  Email and password (--email).
     Signs in to app.bitrise.io with your account credentials, then asks the
     server to mint a fresh Personal Access Token and stores it. The cookie
     session used to mint the token is dropped immediately. Your account must
     have its email verified:

         bitrise auth login --email alice@example.com
         printf '%s' "$PW" | bitrise auth login --email alice@example.com --password-stdin

     The target is web_base_url (default app.bitrise.io), overridable via
     $BITRISE_WEB_BASE_URL or 'bitrise config set web_base_url' — never by a
     per-directory .bitrise-cli.yml, so a repo you merely clone can't
     silently redirect where your password is sent.

The resulting token is written to $XDG_CONFIG_HOME/bitrise/cli/auth.yaml with
0600 permissions and is never echoed (use 'auth status' to verify, 'auth
logout' to clear).

```
bitrise auth login [flags]
```

### Examples

```
  bitrise auth login                                     # browser sign-in (OAuth)
  echo "$BITRISE_PAT" | bitrise auth login --with-token  # paste/pipe a token
  bitrise auth login --email alice@example.com           # email/password
```

### Options

```
      --email string     sign in by email/password and mint a Personal Access Token
  -h, --help             help for login
      --oauth            sign in via the browser (OAuth) and store a managed, auto-refreshing token
      --password-stdin   with --email, read the password from stdin without prompting
      --with-token       read a Personal Access Token from stdin, prompting for it (masked) in a terminal
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

* [bitrise auth](bitrise_auth.md)	 - Manage the Bitrise access token

