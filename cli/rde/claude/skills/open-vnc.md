## open-vnc — show the session's desktop on the user's screen

When the user wants to see the session's desktop, screen, GUI, a running
simulator, emulator, app, or browser, open a VNC viewer on their machine:

```bash
curl -sS -X POST "$URL/open-vnc" -H "Authorization: Bearer $TOKEN"
```

On success the response is `{"opened": true, ...}` and a VNC viewer opens on the
user's machine — let them know it should now be on their screen. This action
takes no parameters; it always targets this session's own desktop.
