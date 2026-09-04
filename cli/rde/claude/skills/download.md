## download — copy a file or directory from the session onto the user's machine

When the user asks to get something from this session onto their own computer —
"download X to my machine", "save that build output locally", "send me the log" —
call download. The file lives here in the session; this copies it to their local
machine (it does not affect anything here).

```bash
curl -sS -X POST "$URL/download" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"remotePath": "/absolute/path/in/this/session"}'
```

- `remotePath` (required): the file or directory here in the session to copy. Use
  an absolute path (run `pwd` if you need to resolve a relative one).
- `localDest` (optional): a destination **folder** on the user's machine — never a
  target filename. Add it **only** when the user says where to put the file, e.g.
  `-d '{"remotePath": "…", "localDest": "~/Downloads"}'`. If you omit it the file
  goes to a temporary folder on their machine — deliberately not their current
  project, so a download can't be committed by accident.

On success the response is `{"downloaded": true, "localPath": "…"}`. **Always tell
the user the returned `localPath`** — that is where the file actually landed, and
when you omitted `localDest` it is a temp folder, not the directory they're
working in.
