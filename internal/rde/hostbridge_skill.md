---
name: rde-host
description: Act on the user's local machine from inside this remote session. Use whenever the user asks for something that has to happen on their own computer rather than in the session — for example seeing something visual from the session on their own screen, or getting a file onto their machine. The actions actually available in this session are listed in the body below; only those work.
allowed-tools: Read(~/.config/rde/**) Bash(curl *)
---

# Acting on the user's local machine

This session can reach the user's local machine through a small "host bridge".
You cannot affect the user's machine by running things here — their screen,
files, and apps live on their own computer, not in this session. To do something
there, call the host bridge.

Before each call, read the bridge address and token. The port can change if the
connection reconnects, so **read the file every time** rather than reusing an
old value:

```bash
cat ~/.config/rde/host-bridge.json
```

It contains `{"url": "...", "token": "..."}`. If the file does not exist, the
host bridge is not available in this session — tell the user and stop.

Then POST to the action you want, authenticating with the token, using the `url`
and `token` you just read:

```bash
curl -sS -X POST "$URL/<action>" -H "Authorization: Bearer $TOKEN"
```

Some actions take parameters as a JSON request body (`-H "Content-Type:
application/json" -d '{…}'`); each action's section below shows its exact form.

Report what the bridge returns — on a non-2xx response the body explains what
went wrong.

Only the actions listed below are available in this session. The bridge ignores
anything else, so do not assume an action exists unless it appears here.
